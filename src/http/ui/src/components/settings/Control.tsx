import { useState } from "react";
import { Button, Grid } from "@mui/material";
import SettingSection from "@common/B4Section";
import { ControlIcon, RestartIcon, RestoreIcon } from "@b4.icons";
import { RestartDialog } from "./RestartDialog";
import { spacing } from "@design";
import { ResetDialog } from "./ResetDialog";

interface ControlSettingsProps {
  loadConfig: () => void;
}

export const ControlSettings = ({ loadConfig }: ControlSettingsProps) => {
  const [saving] = useState(false);
  const [showRestartDialog, setShowRestartDialog] = useState(false);
  const [showResetDialog, setShowResetDialog] = useState(false);

  const handleResetSuccess = () => {
    loadConfig();
  };

  return (
    <SettingSection
      title="Core Controls"
      description="Control core service and config operations"
      icon={<ControlIcon />}
    >
      <Grid container spacing={spacing.lg}>
        <Button
          size="small"
          variant="outlined"
          startIcon={<RestartIcon />}
          onClick={() => setShowRestartDialog(true)}
          disabled={saving}
        >
          Restart B4 System Service
        </Button>
        <Button
          size="small"
          variant="outlined"
          startIcon={<RestoreIcon />}
          onClick={() => setShowResetDialog(true)}
          disabled={saving}
        >
          Reset the configuration to default settings
        </Button>
      </Grid>

      <RestartDialog
        open={showRestartDialog}
        onClose={() => setShowRestartDialog(false)}
      />

      <ResetDialog
        open={showResetDialog}
        onClose={() => setShowResetDialog(false)}
        onSuccess={handleResetSuccess}
      />
    </SettingSection>
  );
};
