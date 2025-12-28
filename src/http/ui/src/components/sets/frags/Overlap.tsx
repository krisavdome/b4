import { AddIcon } from "@b4.icons";
import { ChipList } from "@components/common/ChipList";
import { Alert, AlertDescription } from "@design/components/ui/alert";
import { Button } from "@design/components/ui/button";
import { Field, FieldLabel } from "@design/components/ui/field";
import { Input } from "@design/components/ui/input";
import { Separator } from "@design/components/ui/separator";
import { B4SetConfig } from "@models/config";
import { useState } from "react";

interface OverlapSettingsProps {
  config: B4SetConfig;
  onChange: (
    field: string,
    value: string | boolean | number | string[]
  ) => void;
}

export const OverlapSettings = ({ config, onChange }: OverlapSettingsProps) => {
  const [newDomain, setNewDomain] = useState("");
  const fakeSNIs = config.fragmentation.overlap.fake_snis || [];

  const handleAddDomain = () => {
    const domain = newDomain.trim().toLowerCase();
    if (domain && !fakeSNIs.includes(domain)) {
      onChange("fragmentation.overlap.fake_snis", [...fakeSNIs, domain]);
      setNewDomain("");
    }
  };

  const handleRemoveDomain = (domain: string) => {
    onChange(
      "fragmentation.overlap.fake_snis",
      fakeSNIs.filter((d) => d !== domain)
    );
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAddDomain();
    }
  };

  return (
    <>
      <div className="relative my-4 md:col-span-2 flex items-center">
        <Separator className="absolute inset-0 top-1/2" />
        <span className="text-xs font-medium text-muted-foreground px-2 uppercase bg-card relative mx-auto block w-fit">
          Overlap Strategy
        </span>
      </div>

      <Alert className="m-0">
        <AlertDescription>
          Exploits RFC 793: server keeps FIRST received data for overlapping
          segments. Real SNI sent first (server sees), fake SNI sent second (DPI
          sees).
        </AlertDescription>
      </Alert>

      {/* Visual explanation */}
      <div className="md:col-span-2">
        <div className="p-4 bg-card rounded-md border border-border">
          <p className="text-xs text-muted-foreground mb-2">
            HOW OVERLAP WORKS
          </p>
          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <p className="text-xs text-muted-foreground min-w-20">
                Sent 1st:
              </p>
              <div className="flex gap-1 font-mono text-xs">
                <div className="p-2 bg-accent-secondary rounded border-2 border-secondary">
                  youtube.com (REAL)
                </div>
                <div className="p-2 bg-accent rounded">...rest</div>
              </div>
              <p className="text-xs text-secondary ml-2">→ Server keeps</p>
            </div>
            <div className="flex items-center gap-2">
              <p className="text-xs text-muted-foreground min-w-20">
                Sent 2nd:
              </p>
              <div className="flex gap-1 font-mono text-xs">
                <div className="p-2 bg-accent rounded">Header...</div>
                <div className="p-2 bg-tertiary rounded border-2 border-dashed border-secondary">
                  {fakeSNIs[0] || "ya.ru"}...... (FAKE)
                </div>
              </div>
              <p className="text-xs text-secondary ml-2">
                → DPI sees, server discards
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Fake SNIs editor */}
      <div className="md:col-span-2">
        <p className="text-sm font-semibold mb-2">Fake SNI Domains</p>
        <p className="text-xs text-muted-foreground mb-4">
          Domains injected into overlapping segment. DPI sees these instead of
          real target. Use allowed/unblocked domains.
        </p>
      </div>

      <div>
        <div className="flex gap-2">
          <Field className="flex-1">
            <FieldLabel>Add Domain</FieldLabel>
            <Input
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="e.g., allowed-site.ru"
            />
          </Field>
          <Button
            variant="secondary"
            size="icon"
            onClick={handleAddDomain}
            disabled={!newDomain.trim()}
          >
            <AddIcon className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div>
        <ChipList
          items={fakeSNIs}
          getKey={(d) => d}
          getLabel={(d) => d}
          onDelete={handleRemoveDomain}
          emptyMessage="No domains configured - defaults will be used"
          showEmpty
        />
      </div>

      {fakeSNIs.length === 0 && (
        <Alert variant="destructive" className="m-0 md:col-span-2">
          <AlertDescription>
            Using default domains (ya.ru, vk.com, etc). Add custom domains that
            are known to be unblocked in your region.
          </AlertDescription>
        </Alert>
      )}

      {fakeSNIs.length > 0 && fakeSNIs.length < 3 && (
        <Alert className="m-0 md:col-span-2">
          <AlertDescription>
            Tip: Add more domains for variety. A random one is selected per
            connection.
          </AlertDescription>
        </Alert>
      )}
    </>
  );
};
