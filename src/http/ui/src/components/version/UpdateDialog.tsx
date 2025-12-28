import { forwardRef, useEffect, useState } from "react";

import {
  CheckCircleIcon,
  CloseIcon,
  CloudDownloadIcon,
  DescriptionIcon,
  NewReleaseIcon,
  OpenInNewIcon,
} from "@b4.icons";
import { Alert, AlertDescription } from "@design/components/ui/alert";
import { Badge } from "@design/components/ui/badge";
import { Button } from "@design/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@design/components/ui/dialog";
import { Label } from "@design/components/ui/label";
import { Progress } from "@design/components/ui/progress";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@design/components/ui/select";
import { Separator } from "@design/components/ui/separator";
import { Switch } from "@design/components/ui/switch";
import { cn } from "@design/lib/utils";
import { GitHubRelease, compareVersions } from "@hooks/useGitHubRelease";
import { useSystemUpdate } from "@hooks/useSystemUpdate";
import React from "react";
import ReactMarkdown from "react-markdown";

interface UpdateModalProps {
  open: boolean;
  onClose: () => void;
  onDismiss: () => void;
  currentVersion: string;
  releases: GitHubRelease[];
  includePrerelease: boolean;
  onTogglePrerelease: (include: boolean) => void;
}

const H2Typography = forwardRef<HTMLHeadingElement, React.ComponentProps<"h2">>(
  function H2Typography(props, ref) {
    return (
      <h2 className="text-sm font-extrabold uppercase" ref={ref} {...props} />
    );
  }
);

export const UpdateModal = ({
  open,
  onClose,
  onDismiss,
  currentVersion,
  releases,
  includePrerelease,
  onTogglePrerelease,
}: UpdateModalProps) => {
  const { performUpdate, waitForReconnection } = useSystemUpdate();
  const [updateStatus, setUpdateStatus] = useState<
    "idle" | "updating" | "reconnecting" | "success" | "error"
  >("idle");
  const [updateMessage, setUpdateMessage] = useState("");
  const [selectedVersion, setSelectedVersion] = useState<string>("");

  useEffect(() => {
    if (releases.length > 0 && !selectedVersion) {
      setSelectedVersion(releases[0].tag_name);
    }
  }, [releases, selectedVersion]);

  useEffect(() => {
    if (!open) {
      setUpdateStatus("idle");
      setUpdateMessage("");
    }
  }, [open]);

  const selectedRelease =
    releases.find((r) => r.tag_name === selectedVersion) || releases[0];

  const isDowngrade =
    selectedVersion &&
    compareVersions(`v${currentVersion}`, selectedVersion) > 0;
  const isCurrent = selectedVersion === `v${currentVersion}`;

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const handleUpdate = async () => {
    setUpdateStatus("updating");
    setUpdateMessage("Initiating update...");

    const result = await performUpdate(selectedVersion);
    if (!result || !result.success) {
      setUpdateStatus("error");
      setUpdateMessage(result?.message || "Failed to initiate update.");
      return;
    }

    setUpdateMessage("Update in progress. Waiting for service to restart...");
    setUpdateStatus("reconnecting");

    const reconnected = await waitForReconnection();

    if (reconnected) {
      setUpdateStatus("success");
      setUpdateMessage("Update completed successfully! Refreshing...");
      setTimeout(() => globalThis.window.location.reload(), 5000);
    } else {
      setUpdateStatus("error");
      setUpdateMessage(
        "Update may have completed but service did not restart. Please check manually."
      );
    }
  };

  const isUpdating =
    updateStatus === "updating" || updateStatus === "reconnecting";

  const getDialogProps = () => {
    const base = {
      title: "Version Management",
      subtitle: selectedRelease
        ? `Published on ${formatDate(selectedRelease.published_at)}`
        : "",
      icon: <NewReleaseIcon />,
    };
    switch (updateStatus) {
      case "updating":
      case "reconnecting":
        return {
          ...base,
          title: "Updating B4 Service",
          subtitle: "Please wait...",
        };
      case "success":
        return { ...base, title: "Update Successful", subtitle: "" };
      case "error":
        return { ...base, title: "Update Failed", subtitle: "" };
      default:
        return base;
    }
  };

  const getStatusContent = () => {
    switch (updateStatus) {
      case "updating":
      case "reconnecting":
        return (
          <div className="mb-6">
            <p className="mb-2 text-muted-foreground">{updateMessage}</p>
            <Progress />
          </div>
        );
      case "success":
        return (
          <Alert className="mb-4">
            <CheckCircleIcon className="h-3.5 w-3.5" />
            <AlertDescription>
              <div className="flex items-center gap-2">{updateMessage}</div>
            </AlertDescription>
          </Alert>
        );
      case "error":
        return (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{updateMessage}</AlertDescription>
          </Alert>
        );
      default:
        return null;
    }
  };

  const dialogContent = () => (
    <>
      {getStatusContent()}

      {updateStatus === "idle" && (
        <div className="mb-6">
          <div className="flex flex-row gap-4 items-center mb-4 mt-4">
            <div className="min-w-55">
              <Label>Select Version</Label>
              <Select
                value={selectedVersion}
                onValueChange={(value) => setSelectedVersion(value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select Version" />
                </SelectTrigger>
                <SelectContent>
                  {releases.map((r) => (
                    <SelectItem key={r.tag_name} value={r.tag_name}>
                      {r.tag_name}
                      {r.prerelease && " (pre-release)"}
                      {r.tag_name === `v${currentVersion}` && " (current)"}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center space-x-2">
              <Switch
                checked={includePrerelease}
                onCheckedChange={(checked) => onTogglePrerelease(checked)}
              />
              <Label>Include pre-releases</Label>
            </div>
          </div>
          <div className="flex flex-row gap-2">
            <Badge variant="default" className="text-xs px-1.5 py-0.5">
              {`Current: v${currentVersion}`}
            </Badge>
            {!isCurrent && (
              <Badge
                variant={isDowngrade ? "destructive" : "secondary"}
                className="text-xs px-1.5 py-0.5 font-semibold"
              >
                {isDowngrade ? "Downgrade" : "Upgrade"}
              </Badge>
            )}
            {selectedRelease?.prerelease && (
              <Badge variant="outline" className="text-xs px-1.5 py-0.5">
                Pre-release
              </Badge>
            )}
          </div>
        </div>
      )}

      {selectedRelease && (
        <div className="max-h-100 overflow-auto p-4 bg-background rounded-md border border-border">
          <h6 className="text-secondary mb-4 font-semibold uppercase">
            Release Notes - {selectedRelease.tag_name}
          </h6>
          <div className="text-foreground [&_h1]:text-secondary [&_h1]:mt-4 [&_h1]:mb-2 [&_h2]:text-secondary [&_h2]:mt-4 [&_h2]:mb-2 [&_h3]:text-secondary [&_h3]:mt-4 [&_h3]:mb-2 [&_p]:mb-2 [&_p]:leading-relaxed [&_ul]:pl-6 [&_ul]:mb-2 [&_ol]:pl-6 [&_ol]:mb-2 [&_code]:bg-card [&_code]:text-secondary [&_code]:px-1 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-sm [&_a]:text-secondary">
            <ReactMarkdown components={{ h2: H2Typography }}>
              {selectedRelease.body || "No release notes available."}
            </ReactMarkdown>
          </div>
        </div>
      )}

      <Separator className="my-4" />

      <div className="flex flex-row gap-4 justify-center">
        <Button variant="outline" asChild>
          <a
            href="https://github.com/DanielLavrushin/b4/blob/main/changelog.md"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2"
          >
            <DescriptionIcon className="h-4 w-4" />
            Full Changelog
          </a>
        </Button>
        {selectedRelease && (
          <Button variant="outline" asChild>
            <a
              href={selectedRelease.html_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2"
            >
              <OpenInNewIcon className="h-4 w-4" />
              View on GitHub
            </a>
          </Button>
        )}
      </div>
    </>
  );

  const dialogActions = () => (
    <>
      <Button onClick={onDismiss} variant="ghost" disabled={isUpdating}>
        <CloseIcon className="h-4 w-4 mr-2" />
        Don't Show Again
      </Button>
      <div className="flex-1" />
      {updateStatus === "idle" && (
        <>
          <Button onClick={onClose} variant="outline" disabled={isUpdating}>
            Close
          </Button>
          <Button
            onClick={() => void handleUpdate()}
            variant="default"
            disabled={isUpdating || isCurrent}
            className={cn(
              isDowngrade && "bg-destructive hover:bg-destructive/90"
            )}
          >
            <CloudDownloadIcon className="h-4 w-4 mr-2" />
            {isDowngrade ? "Downgrade" : "Update"}
          </Button>
        </>
      )}
      {updateStatus === "success" && (
        <Button
          variant="default"
          onClick={() => globalThis.window.location.reload()}
        >
          Reload Page
        </Button>
      )}
    </>
  );

  const dialogProps = getDialogProps();

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => !open && (isUpdating ? () => {} : onClose())}
    >
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-md bg-accent text-accent-foreground">
              {dialogProps.icon}
            </div>
            <div className="flex-1">
              <DialogTitle>{dialogProps.title}</DialogTitle>
              {dialogProps.subtitle && (
                <DialogDescription className="mt-1">
                  {dialogProps.subtitle}
                </DialogDescription>
              )}
            </div>
          </div>
        </DialogHeader>
        <div className="py-4">{dialogContent()}</div>
        {dialogActions() && (
          <>
            <Separator />
            <DialogFooter>{dialogActions()}</DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};
