import { Alert, AlertProps } from "@mui/material";
import { colors } from "@design";

export const B4Alert = ({ sx, ...props }: AlertProps) => (
  <Alert
    sx={{
      bgcolor: colors.background.default,
      border: `1px solid ${colors.border.default}`,
      ...sx,
    }}
    {...props}
  />
);
