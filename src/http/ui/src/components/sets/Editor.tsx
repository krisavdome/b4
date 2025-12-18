import { useEffect, useState } from "react";
import { Box, Stack, Button, Paper, CircularProgress } from "@mui/material";

import {
  DomainIcon,
  TcpIcon,
  UdpIcon,
  DnsIcon,
  FragIcon,
  FakingIcon,
  ImportExportIcon,
} from "@b4.icons";

import { B4Dialog, B4Tab, B4Tabs, B4TextField } from "@b4.elements";

import { colors } from "@design";
import {
  B4Config,
  B4SetConfig,
  MAIN_SET_ID,
  SystemConfig,
} from "@models/config";

import { TargetSettings } from "./Target";
import { TcpSettings } from "./Tcp";
import { UdpSettings } from "./Udp";
import { FragmentationSettings } from "./Fragmentation";
import { ImportExportSettings } from "./ImportExport";
import { DnsSettings } from "./Dns";
import { FakingSettings } from "./Faking";
import { SetStats } from "./Manager";

export interface SetEditorProps {
  open: boolean;
  settings: SystemConfig;
  set: B4SetConfig;
  config: B4Config;
  stats?: SetStats;
  isNew: boolean;
  saving: boolean;
  onClose: () => void;
  onSave: (set: B4SetConfig) => void;
}

export const SetEditor = ({
  open,
  set: initialSet,
  config,
  isNew,
  settings,
  stats,
  saving,
  onClose,
  onSave,
}: SetEditorProps) => {
  enum TABS {
    TARGETS = 0,
    TCP,
    UDP,
    DNS,
    FRAGMENTATION,
    FAKING,
    IMPORT_EXPORT,
  }

  const [activeTab, setActiveTab] = useState<TABS>(TABS.TARGETS);
  const [editedSet, setEditedSet] = useState<B4SetConfig | null>(initialSet);

  const mainSet = config.sets.find((s) => s.id === MAIN_SET_ID)!;

  useEffect(() => {
    setEditedSet(initialSet);
    setActiveTab(0);
  }, [initialSet]);

  const handleChange = (
    field: string,
    value: string | number | boolean | string[] | number[] | null | undefined
  ) => {
    if (!editedSet) return;

    const keys = field.split(".");

    if (keys.length === 1) {
      setEditedSet({ ...editedSet, [field]: value });
    } else {
      const newConfig = { ...editedSet };
      let current: Record<string, unknown> = newConfig;

      for (let i = 0; i < keys.length - 1; i++) {
        current[keys[i]] = { ...(current[keys[i]] as object) };
        current = current[keys[i]] as Record<string, unknown>;
      }

      current[keys[keys.length - 1]] = value;
      setEditedSet(newConfig);
    }
  };

  const handleSave = () => {
    if (editedSet) {
      onSave(editedSet);
    }
  };

  const handleApplyImport = (importedSet: B4SetConfig) => {
    setEditedSet(importedSet);
    setActiveTab(TABS.TARGETS);
  };

  if (!editedSet) return null;

  const dialogContent = (
    <Stack spacing={3} sx={{ mt: 2 }}>
      <Paper
        elevation={0}
        sx={{
          bgcolor: colors.background.paper,
          borderRadius: 2,
          border: `1px solid ${colors.border.default}`,
        }}
      >
        <Box sx={{ mt: 2, p: 3 }}>
          <B4TextField
            label="Set Name"
            value={editedSet.name}
            onChange={(e) => handleChange("name", e.target.value)}
            placeholder="e.g., YouTube Bypass, Gaming, Streaming"
            helperText="Give this set a descriptive name"
            required
          />
        </Box>
        {/* Configuration Tabs */}
        <B4Tabs value={activeTab} onChange={(_, v: number) => setActiveTab(v)}>
          <B4Tab icon={<DomainIcon />} label="Targets" />
          <B4Tab icon={<TcpIcon />} label="TCP" />
          <B4Tab icon={<UdpIcon />} label="UDP" />
          <B4Tab icon={<DnsIcon />} label="DNS" />
          <B4Tab icon={<FragIcon />} label="Fragmentation" />
          <B4Tab icon={<FakingIcon />} label="Faking" />
          <B4Tab icon={<ImportExportIcon />} label="Import/Export" />
        </B4Tabs>
      </Paper>
      <Box>
        {/* TCP Settings */}
        <Box hidden={activeTab !== TABS.TCP}>
          <Stack spacing={2}>
            <TcpSettings
              config={editedSet}
              main={mainSet}
              onChange={handleChange}
            />
          </Stack>
        </Box>

        {/* UDP Settings */}
        <Box hidden={activeTab !== TABS.UDP}>
          <Stack spacing={2}>
            <UdpSettings
              config={editedSet}
              main={mainSet}
              onChange={handleChange}
            />
          </Stack>
        </Box>

        {/* DNS Settings */}
        <Box hidden={activeTab !== TABS.DNS}>
          <Stack spacing={2}>
            <DnsSettings
              config={editedSet}
              onChange={handleChange}
              ipv6={config.queue.ipv6}
            />
          </Stack>
        </Box>

        {/* Fragmentation Settings */}
        <Box hidden={activeTab !== TABS.FRAGMENTATION}>
          <Stack spacing={2}>
            <FragmentationSettings config={editedSet} onChange={handleChange} />
          </Stack>
        </Box>

        {/* Faking Settings */}
        <Box hidden={activeTab !== TABS.FAKING}>
          <Stack spacing={2}>
            <FakingSettings config={editedSet} onChange={handleChange} />
          </Stack>
        </Box>

        {/* Target Settings */}
        <Box hidden={activeTab !== TABS.TARGETS}>
          <Stack spacing={2}>
            <TargetSettings
              geo={settings.geo}
              config={editedSet}
              stats={stats}
              onChange={handleChange}
            />
          </Stack>
        </Box>

        {/* Import/Export Settings */}
        <Box hidden={activeTab !== TABS.IMPORT_EXPORT}>
          <Stack spacing={2}>
            <ImportExportSettings
              config={editedSet}
              onImport={handleApplyImport}
            />
          </Stack>
        </Box>
      </Box>
    </Stack>
  );

  return (
    <B4Dialog
      title={isNew ? "Create New Set" : `Edit Set: ${editedSet.name}`}
      open={open}
      onClose={onClose}
      icon={<ImportExportIcon />}
      fullWidth={true}
      maxWidth="lg"
      actions={
        <>
          <Button onClick={onClose} disabled={saving}>
            Cancel
          </Button>
          <Box sx={{ flex: 1 }} />
          <Button
            variant="contained"
            onClick={handleSave}
            disabled={!editedSet.name.trim() || saving}
            sx={{ minWidth: 140 }}
          >
            {saving ? (
              <>
                <CircularProgress size={16} sx={{ mr: 1, color: "inherit" }} />
                Saving...
              </>
            ) : isNew ? (
              "Create Set"
            ) : (
              "Save Changes"
            )}
          </Button>
        </>
      }
    >
      {dialogContent}
    </B4Dialog>
  );
};
