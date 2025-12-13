import { Container, Alert, Stack } from "@mui/material";
import { DiscoveryIcon } from "@b4.icons";
import { DiscoveryRunner } from "./Discovery";
import { colors } from "@design";

export function DiscoveryPage() {
  return (
    <Container
      maxWidth={false}
      sx={{
        height: "100%",
        display: "flex",
        flexDirection: "column",
        overflow: "auto",
        py: 3,
      }}
    >
      <Stack spacing={3}>
        <Alert severity="warning">
          This feature is EXPERIMENTAL and may affect your current
          configuration.
        </Alert>
        <Alert
          severity="info"
          sx={{
            bgcolor: colors.accent.primary,
            border: `1px solid ${colors.secondary}44`,
          }}
          icon={<DiscoveryIcon />}
        >
          <strong>Configuration Discovery:</strong> Automatically test multiple
          configuration presets to find the most effective DPI bypass settings
          for the domains you specify below. B4 will temporarily apply different
          configurations and measure their performance.
        </Alert>
        <DiscoveryRunner />
      </Stack>
    </Container>
  );
}
