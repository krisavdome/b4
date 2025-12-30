import { formatNumber } from "@utils";
import { colors } from "@design";
import { ProtocolChip } from "@common/ProtocolChip";
import { Card } from "@design/components/ui/card";
import { Badge } from "@design/components/ui/badge";
import { Separator } from "@design/components/ui/separator";

interface Connection {
  timestamp: string;
  protocol: "TCP" | "UDP";
  domain: string;
  source: string;
  destination: string;
  is_target: boolean;
}

interface DashboardActivityPanelsProps {
  topDomains: Record<string, number>;
  recentConnections: Connection[];
}

export const DashboardActivityPanels = ({
  topDomains,
  recentConnections,
}: DashboardActivityPanelsProps) => {
  const topDomainsData = Object.entries(topDomains)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 10);

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div className="col-span-1">
        <Card className="p-4 border border-border">
          <h6 className="text-lg font-semibold mb-4 text-foreground">
            Top Domains
          </h6>
          {topDomainsData.length > 0 ? (
            <ul className="space-y-0">
              {topDomainsData.map(([domain, count], index) => (
                <li key={domain}>
                  <div className="flex flex-row justify-between items-center py-2">
                    <p className="text-sm text-foreground">
                      {index + 1}. {domain}
                    </p>
                    <Badge
                      variant="default"
                      style={{
                        backgroundColor: `${colors.accent.primary}`,
                        color: colors.primary,
                      }}
                    >
                      {formatNumber(count)}
                    </Badge>
                  </div>
                  {index < topDomainsData.length - 1 && <Separator />}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-muted-foreground text-center py-8">
              No domain data available yet
            </p>
          )}
        </Card>
      </div>

      <div className="col-span-1">
        <Card className="p-4 h-full border border-border">
          <h6 className="text-lg font-semibold mb-4 text-foreground">
            Recent Activity
          </h6>
          <ul className="max-h-100 overflow-auto space-y-0">
            {recentConnections.map((conn) => (
              <li key={conn.timestamp} className="py-2">
                <div className="flex flex-col gap-1">
                  <div className="flex flex-row gap-2 items-center">
                    <ProtocolChip protocol={conn.protocol} />
                    <p className="text-sm text-foreground">{conn.domain}</p>
                    {conn.is_target && (
                      <Badge
                        variant="default"
                        className="font-semibold bg-green-500/20 text-green-500"
                      >
                        TARGET
                      </Badge>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {conn.source} → {conn.destination} •{" "}
                    {new Date(conn.timestamp).toLocaleTimeString()}
                  </p>
                </div>
              </li>
            ))}
            {recentConnections.length === 0 && (
              <p className="text-muted-foreground text-center py-8">
                No recent connections
              </p>
            )}
          </ul>
        </Card>
      </div>
    </div>
  );
};
