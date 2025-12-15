import { useCallback, useEffect, useState } from "react";
import {
  Container,
  Box,
  Backdrop,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import { useSnackbar } from "@context/SnackbarProvider";
import { SetsManager, SetWithStats } from "./Manager";
import { B4Config } from "@models/Config";
import { colors } from "@design";

export function SetsPage() {
  const { showError } = useSnackbar();
  const [config, setConfig] = useState<
    (B4Config & { sets?: SetWithStats[] }) | null
  >(null);
  const [loading, setLoading] = useState(true);

  const loadConfig = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch("/api/config");
      if (!response.ok) throw new Error("Failed to load");
      const data = (await response.json()) as B4Config & {
        sets?: SetWithStats[];
      };
      setConfig(data);
    } catch {
      showError("Failed to load configuration");
    } finally {
      setLoading(false);
    }
  }, [showError, setLoading]);

  useEffect(() => {
    void loadConfig();
  }, [loadConfig]);

  if (loading || !config) {
    return (
      <Backdrop open sx={{ zIndex: 9999 }}>
        <Stack alignItems="center" spacing={2}>
          <CircularProgress sx={{ color: colors.secondary }} />
          <Typography sx={{ color: colors.text.primary }}>
            Loading...
          </Typography>
        </Stack>
      </Backdrop>
    );
  }

  return (
    <Container
      maxWidth={false}
      sx={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
        py: 3,
      }}
    >
      <Box sx={{ flex: 1, overflow: "auto" }}>
        <SetsManager config={config} onRefresh={() => void loadConfig()} />
      </Box>
    </Container>
  );
}
