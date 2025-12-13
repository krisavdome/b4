import { Chip } from "@mui/material";
import { TcpIcon, UdpIcon } from "@b4.icons";
import { colors } from "@design";

interface ProtocolChipProps {
  protocol: "TCP" | "UDP";
}

export const ProtocolChip = ({ protocol }: ProtocolChipProps) => {
  return (
    <Chip
      label={protocol}
      size="small"
      icon={
        protocol === "TCP" ? (
          <TcpIcon color="primary" />
        ) : (
          <UdpIcon color="secondary" />
        )
      }
      sx={{
        bgcolor: colors.accent.primary,
        color: protocol === "TCP" ? colors.primary : colors.secondary,
        fontWeight: 600,
      }}
    />
  );
};
