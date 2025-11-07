import { useState, useEffect } from "react";

interface GitHubRelease {
  tag_name: string;
  name: string;
  body: string;
  html_url: string;
  published_at: string;
  prerelease: boolean;
}

interface UseGitHubReleaseResult {
  latestRelease: GitHubRelease | null;
  isNewVersionAvailable: boolean;
  isLoading: boolean;
  error: string | null;
  currentVersion: string;
}

const GITHUB_REPO = "DanielLavrushin/b4";
const GITHUB_API_URL = `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`;
const DISMISSED_VERSIONS_KEY = "b4_dismissed_versions";

/**
 * Compares two semantic version strings
 * Returns true if version1 is less than version2
 */
const isVersionLower = (version1: string, version2: string): boolean => {
  // Remove 'v' prefix if present
  const v1 = version1.replace(/^v/, "");
  const v2 = version2.replace(/^v/, "");

  // Handle dev version
  if (v1 === "dev") return true;

  const parts1 = v1.split(".").map(Number);
  const parts2 = v2.split(".").map(Number);

  for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
    const part1 = parts1[i] || 0;
    const part2 = parts2[i] || 0;

    if (part1 < part2) return true;
    if (part1 > part2) return false;
  }

  return false;
};

/**
 * Get list of dismissed version updates from localStorage
 */
const getDismissedVersions = (): string[] => {
  try {
    const dismissed = localStorage.getItem(DISMISSED_VERSIONS_KEY);
    return dismissed ? (JSON.parse(dismissed) as string[]) : [];
  } catch {
    return [];
  }
};

/**
 * Check if a specific version has been dismissed
 */
export const isVersionDismissed = (version: string): boolean => {
  const dismissed = getDismissedVersions();
  return dismissed.includes(version);
};

/**
 * Dismiss a specific version update
 */
export const dismissVersion = (version: string): void => {
  try {
    const dismissed = getDismissedVersions();
    if (!dismissed.includes(version)) {
      dismissed.push(version);
      localStorage.setItem(DISMISSED_VERSIONS_KEY, JSON.stringify(dismissed));
    }
  } catch (error) {
    console.error("Failed to save dismissed version:", error);
  }
};

/**
 * Hook to fetch and manage GitHub release information
 */
export const useGitHubRelease = (): UseGitHubReleaseResult => {
  const [latestRelease, setLatestRelease] = useState<GitHubRelease | null>(
    null
  );
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const currentVersion = import.meta.env.VITE_APP_VERSION || "dev";

  useEffect(() => {
    const fetchLatestRelease = async () => {
      try {
        setIsLoading(true);
        setError(null);

        const response = await fetch(GITHUB_API_URL, {
          headers: {
            Accept: "application/vnd.github.v3+json",
          },
        });

        if (!response.ok) {
          throw new Error(`GitHub API returned ${response.status}`);
        }

        const data: GitHubRelease = (await response.json()) as GitHubRelease;

        // Only set if it's not a prerelease
        if (!data.prerelease) {
          setLatestRelease(data);
        }
      } catch (err) {
        console.error("Failed to fetch GitHub release:", err);
        setError(err instanceof Error ? err.message : "Unknown error");
      } finally {
        setIsLoading(false);
      }
    };

    void fetchLatestRelease();

    // Check for updates every 6 hours
    const interval = setInterval(() => {
      void fetchLatestRelease();
    }, 6 * 60 * 60 * 1000);

    return () => clearInterval(interval);
  }, []);

  const isNewVersionAvailable =
    latestRelease !== null &&
    isVersionLower(currentVersion, latestRelease.tag_name) &&
    !isVersionDismissed(latestRelease.tag_name);

  return {
    latestRelease,
    isNewVersionAvailable,
    isLoading,
    error,
    currentVersion,
  };
};
