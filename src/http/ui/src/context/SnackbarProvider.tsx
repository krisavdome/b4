// src/http/ui/src/context/SnackbarContext.tsx
import {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
} from "react";
import { Snackbar } from "@mui/material";
import { B4Alert } from "@b4.elements";

type Severity = "error" | "warning" | "info" | "success";

interface SnackbarState {
  open: boolean;
  message: string;
  severity: Severity;
}

interface SnackbarContextType {
  showSnackbar: (message: string, severity?: Severity) => void;
  showError: (message: string) => void;
  showSuccess: (message: string) => void;
}

const SnackbarContext = createContext<SnackbarContextType | null>(null);

export function SnackbarProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<SnackbarState>({
    open: false,
    message: "",
    severity: "info",
  });

  const showSnackbar = useCallback(
    (message: string, severity: Severity = "info") => {
      setState({ open: true, message, severity });
    },
    []
  );

  const showError = useCallback(
    (message: string) => showSnackbar(message, "error"),
    [showSnackbar]
  );
  const showSuccess = useCallback(
    (message: string) => showSnackbar(message, "success"),
    [showSnackbar]
  );

  const handleClose = useCallback(() => {
    setState((prev) => ({ ...prev, open: false }));
  }, []);

  return (
    <SnackbarContext.Provider value={{ showSnackbar, showError, showSuccess }}>
      {children}
      <Snackbar
        open={state.open}
        autoHideDuration={6000}
        onClose={handleClose}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
      >
        <B4Alert noWrapper onClose={handleClose} severity={state.severity}>
          {state.message}
        </B4Alert>
      </Snackbar>
    </SnackbarContext.Provider>
  );
}

export function useSnackbar(): SnackbarContextType {
  const context = useContext(SnackbarContext);
  if (!context) {
    throw new Error("useSnackbar must be used within SnackbarProvider");
  }
  return context;
}
