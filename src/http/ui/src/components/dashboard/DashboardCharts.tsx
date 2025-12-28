import { SimpleLineChart } from "./SimpleLineChart";
import { colors } from "@design";
import { Card } from "@design/components/ui/card";

interface DashboardChartsProps {
  connectionRate: { timestamp: number; value: number }[];
  protocolDist: Record<string, number>;
}

export const DashboardCharts = ({ connectionRate }: DashboardChartsProps) => {
  return (
    <div className="grid grid-cols-1 gap-6">
      <div className="col-span-1 lg:col-span-1">
        <Card className="p-4 border border-border">
          <h6 className="text-lg font-semibold mb-4 text-foreground">
            Connection Rate (last 60s)
          </h6>
          <div className="pl-12">
            <SimpleLineChart data={connectionRate} color={colors.secondary} />
          </div>
        </Card>
      </div>
    </div>
  );
};
