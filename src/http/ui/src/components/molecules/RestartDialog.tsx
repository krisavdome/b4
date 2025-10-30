import React, { useState } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Button,
  Alert,
  CircularProgress,
  Stack,
  Typography,
  LinearProgress,
  Box,
} from "@mui/material";
import {
  RestartAlt as RestartIcon,
  CheckCircle as CheckIcon,
  Error as ErrorIcon,
} from "@mui/icons-material";
import { useSystemRestart } from "../../hooks/useSystemRestart";
import { colors } from "../../Theme";

interface RestartDialogProps {
  open: boolean;
  onClose: () => void;
}

type RestartState = "confirm" | "restarting" | "waiting" | "success" | "error";

export const RestartDialog: React.FC<RestartDialogProps> = ({
  open,
  onClose,
}) => {
  const [state, setState] = useState<RestartState>("confirm");
  const [message, setMessage] = useState("");
  const { restart, waitForReconnection, error } = useSystemRestart();

  const handleRestart = async () => {
    setState("restarting");
    setMessage("Initiating restart...");

    const response = await restart();

    if (response && response.success) {
      setState("waiting");
      setMessage("Service is restarting, waiting for reconnection...");

      // Wait for service to come back online
      const reconnected = await waitForReconnection(30);

      if (reconnected) {
        setState("success");
        setMessage("Service restarted successfully!");

        // Auto-close and reload after success
        setTimeout(() => {
          window.location.reload();
        }, 2000);
      } else {
        setState("error");
        setMessage("Service restart timed out. Please check manually.");
      }
    } else {
      setState("error");
      setMessage(error || "Failed to restart service");
    }
  };

  const handleClose = () => {
    if (state !== "restarting" && state !== "waiting") {
      setState("confirm");
      setMessage("");
      onClose();
    }
  };

  const getDialogContent = () => {
    switch (state) {
      case "confirm":
        return (
          <>
            <DialogContent>
              <DialogContentText>
                This will restart the B4 service. The web interface will be
                temporarily unavailable during the restart (approximately 5-10
                seconds).
              </DialogContentText>
              <Alert severity="info" sx={{ mt: 2 }}>
                Your current configuration will be preserved.
              </Alert>
            </DialogContent>
            <DialogActions>
              <Button onClick={handleClose} color="inherit">
                Cancel
              </Button>
              <Button
                onClick={handleRestart}
                variant="contained"
                startIcon={<RestartIcon />}
                sx={{
                  bgcolor: colors.secondary,
                  "&:hover": { bgcolor: colors.primary },
                }}
              >
                Restart Service
              </Button>
            </DialogActions>
          </>
        );

      case "restarting":
      case "waiting":
        return (
          <>
            <DialogContent>
              <Stack spacing={2} alignItems="center" sx={{ py: 3 }}>
                <CircularProgress size={48} sx={{ color: colors.secondary }} />
                <Typography variant="body1">{message}</Typography>
                <Box sx={{ width: "100%" }}>
                  <LinearProgress
                    sx={{
                      bgcolor: colors.accent.primary,
                      "& .MuiLinearProgress-bar": {
                        bgcolor: colors.secondary,
                      },
                    }}
                  />
                </Box>
                <Typography
                  variant="caption"
                  sx={{ color: colors.text.secondary }}
                >
                  Please wait, do not close this window...
                </Typography>
              </Stack>
            </DialogContent>
          </>
        );

      case "success":
        return (
          <>
            <DialogContent>
              <Stack spacing={2} alignItems="center" sx={{ py: 3 }}>
                <CheckIcon sx={{ fontSize: 64, color: colors.secondary }} />
                <Typography variant="h6">{message}</Typography>
                <Typography
                  variant="body2"
                  sx={{ color: colors.text.secondary }}
                >
                  Reloading interface...
                </Typography>
              </Stack>
            </DialogContent>
          </>
        );

      case "error":
        return (
          <>
            <DialogContent>
              <Stack spacing={2} alignItems="center" sx={{ py: 3 }}>
                <ErrorIcon sx={{ fontSize: 64, color: colors.quaternary }} />
                <Typography variant="h6">Restart Failed</Typography>
                <Alert severity="error" sx={{ width: "100%" }}>
                  {message}
                </Alert>
              </Stack>
            </DialogContent>
            <DialogActions>
              <Button onClick={handleClose} variant="contained">
                Close
              </Button>
            </DialogActions>
          </>
        );
    }
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth="sm"
      fullWidth
      disableEscapeKeyDown={state === "restarting" || state === "waiting"}
      PaperProps={{
        sx: {
          bgcolor: colors.background.paper,
          border: `1px solid ${colors.border.default}`,
        },
      }}
    >
      <DialogTitle
        sx={{ borderBottom: `1px solid ${colors.border.light}`, pb: 2 }}
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <RestartIcon sx={{ color: colors.secondary }} />
          <span>Restart B4 Service</span>
        </Stack>
      </DialogTitle>
      {getDialogContent()}
    </Dialog>
  );
};
