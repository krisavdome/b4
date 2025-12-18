import { Grid } from "@mui/material";
import { LogsIcon } from "@b4.icons";
import { B4Section, B4Select, B4Switch, B4TextField } from "@b4.elements";
import { B4Config, LogLevel } from "@models/config";

interface LoggingSettingsProps {
  config: B4Config;
  onChange: (
    field: string,
    value: number | boolean | string | string[]
  ) => void;
}

const LOG_LEVELS: Array<{ value: LogLevel; label: string }> = [
  { value: LogLevel.ERROR, label: "Error" },
  { value: LogLevel.INFO, label: "Info" },
  { value: LogLevel.TRACE, label: "Trace" },
  { value: LogLevel.DEBUG, label: "Debug" },
] as const;

export const LoggingSettings = ({ config, onChange }: LoggingSettingsProps) => {
  return (
    <B4Section
      title="Logging Configuration"
      description="Configure logging behavior and output"
      icon={<LogsIcon />}
    >
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <B4Select
            label="Log Level"
            value={config.system.logging.level}
            options={LOG_LEVELS}
            onChange={(e) =>
              onChange("system.logging.level", Number(e.target.value))
            }
            helperText="Set the verbosity of logging output"
          />
          <B4TextField
            label="Error File Path"
            value={config.system.logging.error_file}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              onChange("system.logging.error_file", e.target.value)
            }
            placeholder="/var/log/b4/errors.log"
            helperText="Full path to error log file"
          />
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <B4Switch
            label="Instant Flush"
            checked={config?.system?.logging?.instaflush}
            onChange={(checked: boolean) =>
              onChange("system.logging.instaflush", Boolean(checked))
            }
            description="Flush logs immediately (may impact performance)"
          />
          <B4Switch
            label="Syslog"
            checked={config?.system?.logging?.syslog}
            onChange={(checked: boolean) =>
              onChange("system.logging.syslog", Boolean(checked))
            }
            description="Enable syslog output"
          />
        </Grid>
      </Grid>
    </B4Section>
  );
};
