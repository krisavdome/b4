import React, { useState, useEffect } from "react";
import {
  Grid,
  Box,
  Chip,
  IconButton,
  Typography,
  Alert,
  Button,
  List,
  ListItem,
  ListItemText,
  Skeleton,
  Tooltip,
  Tabs,
  Tab,
  Stack,
} from "@mui/material";
import {
  Language as LanguageIcon,
  Add as AddIcon,
  Info as InfoIcon,
  Category as CategoryIcon,
  Domain as DomainIcon,
  Block as BlockIcon,
  Security as SecurityIcon,
} from "@mui/icons-material";
import SettingSection from "@molecules/common/B4Section";
import SettingTextField from "@atoms/common/B4TextField";
import SettingAutocomplete from "@atoms/common/B4Autocomplete";
import { colors, button_primary } from "@design";
import { B4Dialog } from "@molecules/common/B4Dialog";
import { B4SetConfig, GeoConfig } from "@models/Config";
import { TargetStatistics } from "@organisms/settings/set/Manager";

interface TargetSettingsProps {
  config: B4SetConfig;
  geo: GeoConfig;
  stats?: TargetStatistics;
  onChange: (field: string, value: string | string[]) => void;
}

interface CategoryPreview {
  category: string;
  total_domains: number;
  preview_count: number;
  preview: string[];
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: Readonly<TabPanelProps>) {
  const { children, value, index, ...other } = props;
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`domain-tabpanel-${index}`}
      aria-labelledby={`domain-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  );
}

export const TargetSettings: React.FC<TargetSettingsProps> = ({
  config,
  onChange,
  geo,
  stats,
}) => {
  const [tabValue, setTabValue] = useState(0);
  const [newBypassDomain, setNewBypassDomain] = useState("");
  const [newBypassCategory, setNewBypassCategory] = useState("");
  const [availableCategories, setAvailableCategories] = useState<string[]>([]);
  const [loadingCategories, setLoadingCategories] = useState(false);

  const [previewDialog, setPreviewDialog] = useState<{
    open: boolean;
    category: string;
    data?: CategoryPreview;
    loading: boolean;
  }>({ open: false, category: "", loading: false });

  useEffect(() => {
    if (geo.sitedat_path) {
      void loadAvailableCategories();
    }
  }, [geo.sitedat_path]);

  const loadAvailableCategories = async () => {
    setLoadingCategories(true);
    try {
      const response = await fetch("/api/geosite");
      if (response.ok) {
        const data = (await response.json()) as { tags: string[] };
        setAvailableCategories(data.tags || []);
      }
    } catch (error) {
      console.error("Failed to load geosite categories:", error);
    } finally {
      setLoadingCategories(false);
    }
  };

  // Bypass domain handlers
  const handleAddBypassDomain = () => {
    if (newBypassDomain.trim()) {
      onChange("domains.sni_domains", [
        ...config.domains.sni_domains,
        newBypassDomain.trim(),
        config.id,
      ]);
      setNewBypassDomain("");
    }
  };

  const handleRemoveBypassDomain = (domain: string) => {
    onChange(
      "domains.sni_domains",
      config.domains.sni_domains.filter((d) => d !== domain)
    );
  };

  const handleAddBypassCategory = (category: string) => {
    if (category && !config.domains.geosite_categories.includes(category)) {
      onChange("domains.geosite_categories", [
        ...config.domains.geosite_categories,
        category,
      ]);
      setNewBypassCategory("");
    }
  };

  const handleRemoveBypassCategory = (category: string) => {
    onChange(
      "domains.geosite_categories",
      config.domains.geosite_categories.filter((c) => c !== category)
    );
  };

  const previewCategory = async (category: string) => {
    setPreviewDialog({ open: true, category, loading: true });
    try {
      const response = await fetch(
        `/api/geosite/category?tag=${encodeURIComponent(category)}`
      );
      if (response.ok) {
        const data = (await response.json()) as CategoryPreview;
        setPreviewDialog((prev) => ({ ...prev, data, loading: false }));
      }
    } catch (error) {
      console.error("Failed to preview category:", error);
      setPreviewDialog((prev) => ({ ...prev, loading: false }));
    }
  };

  return (
    <>
      <Stack spacing={3}>
        <SettingSection
          title="Domain Filtering Configuration"
          description="Configure domain matching for DPI bypass and blocking"
          icon={<LanguageIcon />}
        >
          <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 0 }}>
            <Tabs
              value={tabValue}
              onChange={(_, newValue: number) => setTabValue(newValue)}
              sx={{
                borderBottom: `1px solid ${colors.border.light}`,
                "& .MuiTab-root": {
                  color: colors.text.secondary,
                  textTransform: "none",
                  minHeight: 48,
                  "&.Mui-selected": {
                    color: colors.secondary,
                  },
                },
                "& .MuiTabs-indicator": {
                  bgcolor: colors.secondary,
                },
              }}
            >
              <Tab
                icon={<SecurityIcon />}
                iconPosition="start"
                label={
                  <Box sx={{ display: "flex", alignItems: "center", gap: 2.5 }}>
                    <span>Bypass Domains</span>
                  </Box>
                }
              />
              <Tab
                icon={<BlockIcon />}
                iconPosition="start"
                label={
                  <Box sx={{ display: "flex", alignItems: "center", gap: 2.5 }}>
                    <span>Block Domains</span>
                  </Box>
                }
              />
            </Tabs>
          </Box>
          {/* DPI Bypass Tab */}
          <TabPanel value={tabValue} index={0}>
            <Alert severity="info" sx={{ mb: 2 }}>
              Domains in this list will use DPI bypass techniques
              (fragmentation, faking) when matched.
            </Alert>

            <Grid container spacing={2}>
              {/* Manual Bypass Domains */}
              <Grid size={{ sm: 12, md: 6 }}>
                <Box sx={{ mb: 2 }}>
                  <Typography
                    variant="h6"
                    sx={{
                      display: "flex",
                      alignItems: "center",
                      gap: 1,
                      mb: 2,
                    }}
                  >
                    <DomainIcon /> Manual Bypass Domains
                    <Tooltip title="Add specific domains to bypass DPI. These take priority over GeoSite categories.">
                      <InfoIcon fontSize="small" color="action" />
                    </Tooltip>
                  </Typography>
                  <Box
                    sx={{ display: "flex", gap: 1, alignItems: "flex-start" }}
                  >
                    <SettingTextField
                      label="Add Bypass Domain"
                      value={newBypassDomain}
                      onChange={(e) => setNewBypassDomain(e.target.value)}
                      onKeyDown={(e) => {
                        if (
                          e.key === "Enter" ||
                          e.key === "Tab" ||
                          e.key === ","
                        ) {
                          e.preventDefault();
                          handleAddBypassDomain();
                        }
                      }}
                      helperText="e.g., youtube.com, google.com"
                      placeholder="example.com"
                    />
                    <IconButton
                      onClick={handleAddBypassDomain}
                      sx={{
                        bgcolor: colors.accent.secondary,
                        color: colors.secondary,
                        "&:hover": {
                          bgcolor: colors.accent.secondaryHover,
                        },
                      }}
                    >
                      <AddIcon />
                    </IconButton>
                  </Box>
                  <Box sx={{ mt: 2 }}>
                    <Typography variant="subtitle2" gutterBottom>
                      Active manually added domains
                    </Typography>
                    <Box
                      sx={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: 1,
                        p: 1,
                        border: `1px solid ${colors.border.default}`,
                        borderRadius: 1,
                        bgcolor: colors.background.paper,
                      }}
                    >
                      {config.domains.sni_domains.length === 0 ? (
                        <Typography variant="body2" color="text.secondary">
                          No bypass domains added
                        </Typography>
                      ) : (
                        config.domains.sni_domains.map((domain) => (
                          <Chip
                            key={domain}
                            label={domain}
                            onDelete={() => handleRemoveBypassDomain(domain)}
                            size="small"
                            sx={{
                              bgcolor: colors.accent.primary,
                              color: colors.secondary,
                              "& .MuiChip-deleteIcon": {
                                color: colors.secondary,
                              },
                            }}
                          />
                        ))
                      )}
                    </Box>
                  </Box>
                </Box>
              </Grid>

              {geo.sitedat_path && availableCategories.length > 0 && (
                <Grid size={{ sm: 12, md: 6 }}>
                  <Box sx={{ mb: 2 }}>
                    <Typography
                      variant="h6"
                      sx={{
                        display: "flex",
                        alignItems: "center",
                        gap: 1,
                        mb: 2,
                      }}
                    >
                      <CategoryIcon /> Bypass GeoSite Categories
                      <Tooltip title="Load predefined domain lists from GeoSite database for DPI bypass">
                        <InfoIcon fontSize="small" color="action" />
                      </Tooltip>
                    </Typography>

                    <SettingAutocomplete
                      label="Add Bypass Category"
                      value={newBypassCategory}
                      options={availableCategories}
                      onChange={setNewBypassCategory}
                      onSelect={handleAddBypassCategory}
                      loading={loadingCategories}
                      placeholder="Select or type category"
                      helperText={`${availableCategories.length} categories available`}
                    />

                    {config.domains.geosite_categories.length > 0 && (
                      <Box sx={{ mt: 2 }}>
                        <Typography variant="subtitle2" gutterBottom>
                          Active Bypass Categories
                        </Typography>
                        <Box
                          sx={{
                            display: "flex",
                            flexWrap: "wrap",
                            gap: 1,
                            p: 1,
                            border: `1px solid ${colors.border.default}`,
                            borderRadius: 1,
                            bgcolor: colors.background.paper,
                          }}
                        >
                          {config.domains.geosite_categories.map((category) => (
                            <Chip
                              size="small"
                              key={category}
                              label={
                                <Box
                                  sx={{
                                    display: "flex",
                                    alignItems: "center",
                                  }}
                                >
                                  <span>{category}</span>
                                  {stats?.category_breakdown?.[category] && (
                                    <Typography
                                      component="span"
                                      variant="caption"
                                      sx={{
                                        cursor: "pointer",
                                        bgcolor: "action.selected",
                                        px: 0.5,
                                        ml: 0.5,
                                        borderRadius: 1,
                                      }}
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        void previewCategory(category);
                                      }}
                                    >
                                      {stats.category_breakdown[category]}
                                    </Typography>
                                  )}
                                </Box>
                              }
                              onDelete={() =>
                                handleRemoveBypassCategory(category)
                              }
                              sx={{
                                bgcolor: colors.accent.primary,
                                color: colors.secondary,
                                "& .MuiChip-deleteIcon": {
                                  color: colors.secondary,
                                },
                              }}
                            />
                          ))}
                        </Box>
                      </Box>
                    )}
                  </Box>
                </Grid>
              )}
            </Grid>
          </TabPanel>
        </SettingSection>
      </Stack>

      {/* Preview Dialog */}
      <B4Dialog
        title={`${previewDialog.category.toUpperCase()}`}
        subtitle="Category Preview"
        icon={<CategoryIcon />}
        open={previewDialog.open}
        onClose={() =>
          setPreviewDialog({ open: false, category: "", loading: false })
        }
        actions={
          <Button
            variant="contained"
            onClick={() =>
              setPreviewDialog({ open: false, category: "", loading: false })
            }
            sx={{
              ...button_primary,
            }}
          >
            Close
          </Button>
        }
      >
        <>
          {(() => {
            if (previewDialog.loading) {
              return (
                <Box sx={{ p: 2 }}>
                  <Skeleton variant="text" />
                  <Skeleton variant="text" />
                  <Skeleton variant="text" />
                </Box>
              );
            } else if (previewDialog.data) {
              return (
                <>
                  <Alert severity="info" sx={{ mb: 2 }}>
                    Total domains in category:{" "}
                    {previewDialog.data.total_domains}
                    {previewDialog.data.total_domains >
                      previewDialog.data.preview_count &&
                      ` (showing first ${previewDialog.data.preview_count})`}
                  </Alert>
                  <List dense sx={{ maxHeight: 600, overflow: "auto" }}>
                    {previewDialog.data.preview.map((domain) => (
                      <ListItem key={domain}>
                        <ListItemText primary={domain} />
                      </ListItem>
                    ))}
                  </List>
                </>
              );
            } else {
              return (
                <Alert severity="error">Failed to load category preview</Alert>
              );
            }
          })()}
        </>
      </B4Dialog>
    </>
  );
};
