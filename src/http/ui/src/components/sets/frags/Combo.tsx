import { Alert, AlertDescription } from "@design/components/ui/alert";
import { Badge } from "@design/components/ui/badge";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldLabel,
  FieldTitle,
} from "@design/components/ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@design/components/ui/select";
import { Separator } from "@design/components/ui/separator";
import { Slider } from "@design/components/ui/slider";
import { Switch } from "@design/components/ui/switch";
import { B4SetConfig, ComboShuffleMode } from "@models/config";

interface ComboSettingsProps {
  config: B4SetConfig;
  onChange: (field: string, value: string | boolean | number) => void;
}

const shuffleModeOptions: { label: string; value: ComboShuffleMode }[] = [
  { label: "Middle Only", value: "middle" },
  { label: "Full Shuffle", value: "full" },
  { label: "Reverse Order", value: "reverse" },
];

export const ComboSettings = ({ config, onChange }: ComboSettingsProps) => {
  const combo = config.fragmentation.combo;
  const middleSni = config.fragmentation.middle_sni;

  const enabledSplits = [
    combo.first_byte_split && "1st byte",
    combo.extension_split && "ext",
    middleSni && "SNI",
  ].filter(Boolean);

  return (
    <>
      <div className="relative my-4 md:col-span-2 flex items-center">
        <Separator className="absolute inset-0 top-1/2" />
        <span className="text-xs font-medium text-muted-foreground px-2 uppercase bg-card relative mx-auto block w-fit">
          Combo Strategy
        </span>
      </div>

      <div>
        <Alert>
          <AlertDescription>
            Combo combines multiple split points and sends segments out of order
            with timing jitter to confuse stateful DPI.
          </AlertDescription>
        </Alert>
      </div>

      {/* Split Points */}
      <div>
        <p className="text-xs font-semibold mb-2 text-muted-foreground">
          Split Points
        </p>
      </div>

      <div>
        <label htmlFor="switch-combo-first-byte">
          <Field orientation="horizontal" className="has-[>[data-state=checked]]:bg-primary/5 dark:has-[>[data-state=checked]]:bg-primary/10 has-[>[data-checked]]:bg-primary/5 dark:has-[>[data-checked]]:bg-primary/10 p-2">
            <FieldContent>
              <FieldTitle>First Byte</FieldTitle>
              <FieldDescription>
                Split after 1st byte (timing desync)
              </FieldDescription>
            </FieldContent>
            <Switch
              id="switch-combo-first-byte"
              checked={combo.first_byte_split}
              onCheckedChange={(checked: boolean) =>
                onChange("fragmentation.combo.first_byte_split", checked)
              }
            />
          </Field>
        </label>
      </div>

      <div>
        <label htmlFor="switch-combo-extension">
          <Field orientation="horizontal" className="has-[>[data-state=checked]]:bg-primary/5 dark:has-[>[data-state=checked]]:bg-primary/10 has-[>[data-checked]]:bg-primary/5 dark:has-[>[data-checked]]:bg-primary/10 p-2">
            <FieldContent>
              <FieldTitle>Extension Split</FieldTitle>
              <FieldDescription>Split before SNI extension</FieldDescription>
            </FieldContent>
            <Switch
              id="switch-combo-extension"
              checked={combo.extension_split}
              onCheckedChange={(checked: boolean) =>
                onChange("fragmentation.combo.extension_split", checked)
              }
            />
          </Field>
        </label>
      </div>

      <div>
        <label htmlFor="switch-combo-sni">
          <Field orientation="horizontal" className="has-[>[data-state=checked]]:bg-primary/5 dark:has-[>[data-state=checked]]:bg-primary/10 has-[>[data-checked]]:bg-primary/5 dark:has-[>[data-checked]]:bg-primary/10 p-2">
            <FieldContent>
              <FieldTitle>SNI Split</FieldTitle>
              <FieldDescription>Split in middle of SNI hostname</FieldDescription>
            </FieldContent>
            <Switch
              id="switch-combo-sni"
              checked={middleSni}
              onCheckedChange={(checked: boolean) =>
                onChange("fragmentation.middle_sni", checked)
              }
            />
          </Field>
        </label>
      </div>

      {/* Visual */}
      <div className="md:col-span-3">
        <div className="p-4 bg-card rounded-md border border-border">
          <p className="text-xs text-muted-foreground mb-2">
            SEGMENT VISUALIZATION
          </p>
          <div className="flex gap-1 font-mono text-xs flex-wrap">
            {combo.first_byte_split && (
              <div className="p-2 bg-tertiary rounded text-center min-w-10">
                1B
              </div>
            )}
            {combo.extension_split && (
              <div className="p-2 bg-accent rounded text-center flex-1 min-w-15">
                Pre-SNI Ext
              </div>
            )}
            {middleSni && (
              <>
                <div className="p-2 bg-accent-secondary rounded text-center min-w-12.5">
                  SNI₁
                </div>
                <div className="p-2 bg-accent-secondary rounded text-center min-w-12.5">
                  SNI₂
                </div>
              </>
            )}
            <div className="p-2 bg-quaternary rounded text-center flex-1 min-w-15">
              Rest...
            </div>
          </div>
          <p className="text-xs text-muted-foreground mt-2">
            {enabledSplits.length > 0
              ? `Active splits: ${enabledSplits.join(" → ")} → creates ${
                  enabledSplits.length + 1
                } segments`
              : "No splits enabled - packet sent as single segment"}
          </p>
        </div>
      </div>

      {/* Shuffle Mode */}
      <div>
        <Field>
          <FieldLabel>Shuffle Mode</FieldLabel>
          <Select
            value={combo.shuffle_mode}
            onValueChange={(value) =>
              onChange("fragmentation.combo.shuffle_mode", value as string)
            }
          >
            <SelectTrigger>
              <SelectValue placeholder="Select shuffle mode" />
            </SelectTrigger>
            <SelectContent>
              {shuffleModeOptions.map((option) => (
                <SelectItem key={option.value} value={option.value.toString()}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldDescription>
            How to reorder segments before sending
          </FieldDescription>
        </Field>
      </div>

      <div>
        <Alert className="my-0">
          <AlertDescription>
            {combo.shuffle_mode === "middle" &&
              "Middle: Keep first & last in place, shuffle middle segments"}
            {combo.shuffle_mode === "full" &&
              "Full: Randomly shuffle all segments"}
            {combo.shuffle_mode === "reverse" &&
              "Reverse: Send segments in reverse order"}
          </AlertDescription>
        </Alert>
      </div>

      <div className="relative my-4 md:col-span-2 flex items-center">
        <Separator className="absolute inset-0 top-1/2" />
        <span className="text-xs font-medium text-muted-foreground px-2 uppercase bg-card relative mx-auto block w-fit">
          Timing Settings
        </span>
      </div>

      <div>
        <Field className="w-full space-y-2">
          <div className="flex items-center justify-between">
            <FieldLabel className="text-sm font-medium">
              First Segment Delay
            </FieldLabel>
            <Badge variant="secondary" className="font-semibold">
              {combo.first_delay_ms} ms
            </Badge>
          </div>
          <Slider
            value={[combo.first_delay_ms]}
            onValueChange={(values) =>
              onChange("fragmentation.combo.first_delay_ms", values[0])
            }
            min={10}
            max={500}
            step={10}
            className="w-full"
          />
          <FieldDescription>Delay after first segment (ms)</FieldDescription>
        </Field>
      </div>

      <div>
        <Field className="w-full space-y-2">
          <div className="flex items-center justify-between">
            <FieldLabel className="text-sm font-medium">Jitter Max</FieldLabel>
            <Badge variant="secondary" className="font-semibold">
              {combo.jitter_max_us} μs
            </Badge>
          </div>
          <Slider
            value={[combo.jitter_max_us]}
            onValueChange={(values) =>
              onChange("fragmentation.combo.jitter_max_us", values[0])
            }
            min={100}
            max={10000}
            step={100}
            className="w-full"
          />
          <FieldDescription>
            Max random delay between other segments (μs)
          </FieldDescription>
        </Field>
      </div>

      {!combo.first_byte_split && !combo.extension_split && !middleSni && (
        <Alert variant="destructive" className="md:col-span-2">
          <AlertDescription>
            No split points enabled. Enable at least one for Combo to work
            effectively.
          </AlertDescription>
        </Alert>
      )}
    </>
  );
};
