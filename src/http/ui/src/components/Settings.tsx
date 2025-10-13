import React from "react";
import {
  Container,
  Paper,
  Typography,
  Box,
  TextField,
  Switch,
  FormControlLabel,
  Divider,
  Stack,
} from "@mui/material";

export default function Settings() {
  const [maxLines, setMaxLines] = React.useState("1000");
  const [autoReconnect, setAutoReconnect] = React.useState(true);
  const [timestampFormat, setTimestampFormat] = React.useState("ISO");

  return (
    <Container
      maxWidth="md"
      sx={{
        flex: 1,
        py: 3,
        px: 3,
        display: "flex",
        flexDirection: "column",
        overflow: "auto",
      }}
    >
      <Paper elevation={0} variant="outlined" sx={{ p: 4 }}>
        <Typography variant="h5" gutterBottom sx={{ color: "#F5AD18", mb: 3 }}>
          Settings
        </Typography>

        <Stack spacing={3}>
          <Box>
            <Typography
              variant="subtitle1"
              gutterBottom
              sx={{ fontWeight: 600 }}
            >
              Display Options
            </Typography>
            <Divider sx={{ mb: 2, borderColor: "rgba(245, 173, 24, 0.24)" }} />

            <TextField
              label="Max Lines to Keep"
              value={maxLines}
              onChange={(e) => setMaxLines(e.target.value)}
              size="small"
              type="number"
              fullWidth
              sx={{ mb: 2 }}
              helperText="Maximum number of log lines to store in memory"
            />

            <TextField
              label="Timestamp Format"
              value={timestampFormat}
              onChange={(e) => setTimestampFormat(e.target.value)}
              size="small"
              fullWidth
              helperText="Format for displaying timestamps (e.g., ISO, UTC, Local)"
            />
          </Box>

          <Box>
            <Typography
              variant="subtitle1"
              gutterBottom
              sx={{ fontWeight: 600 }}
            >
              Connection Options
            </Typography>
            <Divider sx={{ mb: 2, borderColor: "rgba(245, 173, 24, 0.24)" }} />

            <FormControlLabel
              control={
                <Switch
                  checked={autoReconnect}
                  onChange={(e) => setAutoReconnect(e.target.checked)}
                  sx={{
                    "& .MuiSwitch-switchBase.Mui-checked": {
                      color: "#F5AD18",
                    },
                    "& .MuiSwitch-switchBase.Mui-checked + .MuiSwitch-track": {
                      backgroundColor: "#F5AD18",
                    },
                  }}
                />
              }
              label="Auto-reconnect on connection loss"
            />
          </Box>

          <Box>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 4 }}>
              B4: ByeBye BigBro - Your log viewer
            </Typography>
          </Box>
        </Stack>
      </Paper>
    </Container>
  );
}
