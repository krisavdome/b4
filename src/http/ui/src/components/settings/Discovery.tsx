import { useState } from "react";
import { Box, Grid, IconButton } from "@mui/material";
import { AddIcon, DiscoveryIcon } from "@b4.icons";
import { B4Config } from "@models/config";
import { colors } from "@design";
import {
  B4Slider,
  B4Section,
  B4TextField,
  B4FormHeader,
  B4ChipList,
} from "@b4.elements";

interface CheckerSettingsProps {
  config: B4Config;
  onChange: (
    field: string,
    value: string | boolean | number | string[]
  ) => void;
}

export const CheckerSettings = ({ config, onChange }: CheckerSettingsProps) => {
  const [newDns, setNewDns] = useState("");

  const handleAddDns = () => {
    if (newDns.trim()) {
      const current = config.system.checker.reference_dns || [];
      if (!current.includes(newDns.trim())) {
        onChange("system.checker.reference_dns", [...current, newDns.trim()]);
      }
      setNewDns("");
    }
  };

  const handleRemoveDns = (dns: string) => {
    const current = config.system.checker.reference_dns || [];
    onChange(
      "system.checker.reference_dns",
      current.filter((s) => s !== dns)
    );
  };

  return (
    <B4Section
      title="Testing Configuration"
      description="Configure testing behavior and output"
      icon={<DiscoveryIcon />}
    >
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 6 }}>
          <B4Slider
            label="Discovery Timeout"
            value={config.system.checker.discovery_timeout || 5}
            onChange={(value) =>
              onChange("system.checker.discovery_timeout", value)
            }
            min={3}
            max={30}
            step={1}
            valueSuffix=" sec"
            helperText="Timeout per preset during discovery"
          />
        </Grid>
        <Grid size={{ xs: 12, lg: 6 }}>
          <B4Slider
            label="Config Propagation Delay"
            value={config.system.checker.config_propagate_ms || 1500}
            onChange={(value) =>
              onChange("system.checker.config_propagate_ms", value)
            }
            min={500}
            max={5000}
            step={100}
            valueSuffix=" ms"
            helperText="Delay for config to propagate to workers (increase on slow devices)"
          />
        </Grid>
        <Grid size={{ xs: 12, lg: 6 }}>
          <B4TextField
            label="Reference Domain"
            value={config.system.checker.reference_domain || "yandex.ru"}
            onChange={(e) =>
              onChange("system.checker.reference_domain", e.target.value)
            }
            placeholder="yandex.ru"
            helperText="Fast domain to measure your network baseline speed"
          />
        </Grid>

        <B4FormHeader label="DNS Configuration" />
        <Grid size={{ xs: 12, md: 6 }}>
          <Box sx={{ display: "flex", gap: 1, alignItems: "flex-start" }}>
            <B4TextField
              label="Add DNS Server"
              value={newDns}
              onChange={(e) => setNewDns(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleAddDns();
                }
              }}
              placeholder="e.g., 8.8.8.8"
              helperText="Additional DNS servers to test"
            />
            <IconButton
              onClick={handleAddDns}
              sx={{
                bgcolor: colors.accent.secondary,
                color: colors.secondary,
                "&:hover": { bgcolor: colors.accent.secondaryHover },
              }}
            >
              <AddIcon />
            </IconButton>
          </Box>
        </Grid>
        <B4ChipList
          items={config.system.checker.reference_dns || []}
          getKey={(d) => d}
          getLabel={(d) => d}
          onDelete={handleRemoveDns}
          title="Active DNS servers to test:"
          gridSize={{ xs: 12, md: 6 }}
        />
      </Grid>
    </B4Section>
  );
};
