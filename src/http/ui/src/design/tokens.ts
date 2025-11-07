export const colors = {
  primary: "#9E1C60" as string,
  secondary: "#F5AD18" as string,
  tertiary: "#811844" as string,
  quaternary: "#561530" as string,
  background: {
    default: "#1a0e15" as string,
    paper: "#1f1218" as string,
    dark: "#0f0a0e" as string,
    control: "rgba(31, 18, 24, 0.6)" as string,
  },
  text: {
    primary: "#ffe8f4" as string,
    secondary: "#f8d7e9" as string,
  },
  border: {
    default: "rgba(245, 173, 24, 0.24)" as string,
    light: "rgba(245, 173, 24, 0.12)" as string,
    medium: "rgba(245, 173, 24, 0.24)" as string,
    strong: "rgba(245, 173, 24, 0.5)" as string,
  },
  accent: {
    primary: "rgba(158, 28, 96, 0.2)" as string,
    primaryHover: "rgba(158, 28, 96, 0.3)" as string,
    primaryStrong: "rgba(158, 28, 96, 0.1)" as string,
    secondary: "rgba(245, 173, 24, 0.2)" as string,
    secondaryHover: "rgba(245, 173, 24, 0.1)" as string,
    tertiary: "rgba(129, 24, 68, 0.2)" as string,
  },
} as const;

export const spacing = {
  xs: 0.5 as number,
  sm: 1 as number,
  md: 2 as number,
  lg: 3 as number,
  xl: 4 as number,
  xxl: 6 as number,
} as const;

export const radius = {
  sm: 1 as number,
  md: 2 as number,
  lg: 3 as number,
  xl: 4 as number,
} as const;

export const typography = {
  sizes: {
    xs: "0.65rem" as string,
    sm: "0.75rem" as string,
    md: "0.875rem" as string,
    lg: "1rem" as string,
    xl: "1.25rem" as string,
  },
  weights: {
    regular: 400 as number,
    medium: 500 as number,
    semibold: 600 as number,
    bold: 700 as number,
  },
} as const;

export const button_primary = {
  bgcolor: colors.primary,
  "&:hover": { bgcolor: colors.secondary },
} as const;

export const button_secondary = {
  color: colors.text.secondary,
  "&:hover": {
    bgcolor: colors.accent.primaryHover,
  },
} as const;
