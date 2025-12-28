import { NetworkIcon } from "@b4.icons";
import { Badge } from "@design/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@design/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldLabel,
} from "@design/components/ui/field";
import { Input } from "@design/components/ui/input";
import { Slider } from "@design/components/ui/slider";
import { B4Config } from "@models/config";

interface NetworkSettingsProps {
  config: B4Config;
  onChange: (field: string, value: number) => void;
}

export const NetworkSettings = ({ config, onChange }: NetworkSettingsProps) => (
  <Card>
    <CardHeader>
      <div className="flex items-center gap-2">
        <NetworkIcon className="h-5 w-5" />
        <CardTitle>Network Configuration</CardTitle>
      </div>
      <CardDescription>
        Configure netfilter queue and network processing parameters
      </CardDescription>
    </CardHeader>
    <CardContent>
      <div className="mb-4">
        <FieldLabel className="text-base font-semibold mb-4">
          Queue Settings
        </FieldLabel>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field>
            <FieldLabel>Queue Start Number</FieldLabel>
            <Input
              type="number"
              value={config.queue.start_num}
              onChange={(e) =>
                onChange("queue.start_num", Number(e.target.value))
              }
            />
            <FieldDescription>
              Netfilter queue number (0-65535)
            </FieldDescription>
          </Field>
          <Field className="w-full space-y-2">
            <div className="flex items-center justify-between">
              <FieldLabel className="text-sm font-medium">
                Worker Threads
              </FieldLabel>
              <Badge variant="secondary" className="font-semibold">
                {config.queue.threads}
              </Badge>
            </div>
            <Slider
              value={[config.queue.threads]}
              onValueChange={(values) => onChange("queue.threads", values[0])}
              min={1}
              max={16}
              step={1}
              className="w-full"
            />
            <FieldDescription>
              Number of worker threads for processing packets simultaneously
              (default 4)
            </FieldDescription>
          </Field>
        </div>
      </div>
    </CardContent>
  </Card>
);
