import { Container, Alert, Stack, Tabs, Tab } from "@mui/material";
import { useState } from "react";
import { TestRunner } from "@organisms/check/Runner";
import { DiscoveryRunner } from "@organisms/check/Discovery";
import { colors } from "@design";

export default function Test() {
  const [activeTab, setActiveTab] = useState(0);

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
        <Tabs
          value={activeTab}
          onChange={(_, newValue: number) => setActiveTab(newValue)}
          sx={{
            borderBottom: `1px solid ${colors.border.default}`,
            "& .MuiTab-root": {
              color: colors.text.secondary,
              "&.Mui-selected": {
                color: colors.secondary,
              },
            },
          }}
        >
          <Tab label="Quick Test" />
          <Tab label="Discovery" />
        </Tabs>

        {activeTab === 0 && (
          <>
            <Alert
              severity="info"
              sx={{
                bgcolor: colors.accent.primary,
                border: `1px solid ${colors.secondary}44`,
              }}
            >
              <strong>Quick Test:</strong> Test your current configuration
              against the domains you specify below. This validates your
              existing DPI bypass settings without making any changes.
            </Alert>
            <TestRunner />
          </>
        )}

        {activeTab === 1 && (
          <>
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
            >
              <strong>Configuration Discovery:</strong> Automatically test
              multiple configuration presets to find the most effective DPI
              bypass settings for the domains you specify below. B4 will
              temporarily apply different configurations and measure their
              performance.
            </Alert>
            <DiscoveryRunner />
          </>
        )}
      </Stack>
    </Container>
  );
}
