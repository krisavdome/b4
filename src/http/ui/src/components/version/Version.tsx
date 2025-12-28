import { GitHubIcon, DescriptionIcon } from "@b4.icons";
import { Separator } from "@design/components/ui/separator";
import { dismissVersion, useGitHubRelease } from "@hooks/useGitHubRelease";
import { useState } from "react";
import { VersionBadge } from "./Badge";
import { UpdateModal } from "./UpdateDialog";

export default function Version() {
  const [updateModalOpen, setUpdateModalOpen] = useState(false);
  const {
    releases,
    latestRelease,
    isNewVersionAvailable,
    isLoading,
    currentVersion,
    includePrerelease,
    setIncludePrerelease,
  } = useGitHubRelease();

  const handleVersionClick = () => {
    setUpdateModalOpen(true);
  };

  const handleDismissUpdate = () => {
    if (latestRelease) {
      dismissVersion(latestRelease.tag_name);
    }
    setUpdateModalOpen(false);
  };

  return (
    <>
      <div className="py-4">
        <Separator className="mb-4" />
        <div className="flex flex-col items-center gap-3">
          <div className="flex flex-col items-center gap-2 w-full">
            <a
              href="https://github.com/daniellavrushin/b4"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-xs text-primary underline-offset-4 hover:underline transition-colors"
            >
              <GitHubIcon className="h-4 w-4" />
              <span>DanielLavrushin/b4</span>
            </a>
            <a
              href="https://daniellavrushin.github.io/b4/"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-xs text-primary underline-offset-4 hover:underline transition-colors"
            >
              <DescriptionIcon className="h-4 w-4" />
              <span>Documentation</span>
            </a>
          </div>
          <VersionBadge
            version={currentVersion}
            hasUpdate={isNewVersionAvailable}
            isLoading={isLoading}
            onClick={handleVersionClick}
          />
        </div>
      </div>

      <UpdateModal
        open={updateModalOpen}
        onClose={() => setUpdateModalOpen(false)}
        onDismiss={handleDismissUpdate}
        currentVersion={currentVersion}
        releases={releases}
        includePrerelease={includePrerelease}
        onTogglePrerelease={setIncludePrerelease}
      />
    </>
  );
}
