import { ImportExportIcon, CheckIcon, RefreshIcon } from "@b4.icons";
import { Alert, AlertDescription } from "@design/components/ui/alert";
import { Button } from "@design/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@design/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldLabel,
} from "@design/components/ui/field";
import { Textarea } from "@design/components/ui/textarea";
import { useEffect, useState } from "react";

import { B4SetConfig } from "@models/config";

interface ImportExportSettingsProps {
  config: B4SetConfig;
  onImport: (importedConfig: B4SetConfig) => void;
}

export const ImportExportSettings = ({
  config,
  onImport,
}: ImportExportSettingsProps) => {
  const [jsonValue, setJsonValue] = useState("");
  const [originalJson, setOriginalJson] = useState("");
  const [validationError, setValidationError] = useState("");
  const [validationSuccess, setValidationSuccess] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    const formatted = JSON.stringify(config, null, 0);
    setJsonValue(formatted);
    setOriginalJson(formatted);
    setValidationError("");
    setValidationSuccess(false);
    setHasChanges(false);
  }, [config]);

  const handleJsonChange = (value: string) => {
    setJsonValue(value);
    setHasChanges(value !== originalJson);
    setValidationError("");
    setValidationSuccess(false);
  };

  const handleValidate = () => {
    try {
      const parsed = JSON.parse(jsonValue) as B4SetConfig;

      // Validate required fields
      if (
        !parsed.name ||
        !parsed.tcp ||
        !parsed.udp ||
        !parsed.fragmentation ||
        !parsed.faking ||
        !parsed.targets
      ) {
        setValidationError(
          "Invalid set configuration: missing required fields"
        );
        setValidationSuccess(false);
        return null;
      }

      setValidationError("");
      setValidationSuccess(true);
      return parsed;
    } catch (error) {
      setValidationError(
        error instanceof Error ? error.message : "Invalid JSON format"
      );
      setValidationSuccess(false);
      return null;
    }
  };

  const handleApply = () => {
    const validated = handleValidate();
    if (validated) {
      // Preserve the original ID
      validated.id = config.id;
      onImport(validated);
    }
  };

  const handleReset = () => {
    setJsonValue(originalJson);
    setHasChanges(false);
    setValidationError("");
    setValidationSuccess(false);
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-md bg-accent text-accent-foreground">
            <ImportExportIcon />
          </div>
          <div className="flex-1">
            <CardTitle>Import/Export Set configuration</CardTitle>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <Alert className="mb-4">
          <AlertDescription>
            You can export the current set configuration as JSON, or import a
            new configuration by pasting valid JSON below.
          </AlertDescription>
        </Alert>
        <div className="flex flex-col gap-4">
          <Field>
            <FieldLabel>Set Configuration JSON</FieldLabel>
            <Textarea
              value={jsonValue}
              onChange={(e) => handleJsonChange(e.target.value)}
              rows={10}
              spellCheck={false}
            />
            <FieldDescription>
              Edit directly or paste a configuration. Changes must be applied to
              take effect.
            </FieldDescription>
          </Field>

          {validationError && (
            <Alert variant="destructive">
              <AlertDescription>{validationError}</AlertDescription>
            </Alert>
          )}

          <div className="flex gap-4">
            <Button
              variant="outline"
              onClick={handleReset}
              disabled={!hasChanges}
            >
              <RefreshIcon className="h-4 w-4 mr-2" />
              Reset
            </Button>
            <div className="flex-1" />
            <div className="flex items-center gap-2">
              {validationSuccess && !validationError && (
                <CheckIcon className="h-4 w-4 text-primary" />
              )}
              <Button
                variant="outline"
                onClick={handleValidate}
                className="hover:bg-input/50 hover:text-foreground"
              >
                Validate
              </Button>
            </div>
            <Button
              onClick={handleApply}
              disabled={!hasChanges}
            >
              Apply Changes
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
