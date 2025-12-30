import { ParsedLog } from "@b4.connections";
import { AddIcon } from "@b4.icons";
import { ProtocolChip } from "@common/ProtocolChip";
import { SortableTableCell, SortDirection } from "@common/SortableTableCell";
import { colors } from "@design";
import { Badge } from "@design/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@design/components/ui/tooltip";
import { cn } from "@design/lib/utils";
import { asnStorage } from "@utils";
import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react";

export type SortColumn =
  | "timestamp"
  | "set"
  | "protocol"
  | "domain"
  | "source"
  | "destination";

interface DomainsTableProps {
  data: ParsedLog[];
  sortColumn: SortColumn | null;
  sortDirection: SortDirection;
  onSort: (column: SortColumn) => void;
  onDomainClick: (domain: string) => void;
  onIpClick: (ip: string) => void;
  onScrollStateChange: (isAtBottom: boolean) => void;
}

const ROW_HEIGHT = 41;
const OVERSCAN = 5;

const TableRowMemo = memo<{
  log: ParsedLog;
  onDomainClick: (domain: string) => void;
  onIpClick: (ip: string) => void;
}>(
  ({ log, onDomainClick, onIpClick }) => {
    const asnName = useMemo(() => {
      if (!log.destination) return null;
      const asn = asnStorage.findAsnForIp(log.destination);
      return asn?.name || null;
    }, [log.destination]);

    return (
      <tr
        className={cn(
          "h-10.25 hover:bg-accent transition-colors",
          `hover:bg-[${colors.accent.primaryStrong}]`
        )}
      >
        <td className="text-muted-foreground font-mono text-xs border-b border-border py-2 px-4">
          {log.timestamp.split(" ")[1]}
        </td>
        <td className="border-b border-border py-2 px-4">
          <ProtocolChip protocol={log.protocol} />
        </td>
        <td className="border-b border-border py-2 px-4">
          {(log.ipSet || log.hostSet) && (
            <Badge variant="secondary">{log.ipSet || log.hostSet}</Badge>
          )}
        </td>
        <td
          className={cn(
            "text-foreground border-b border-border py-2 px-4",
            log.domain &&
              !log.hostSet &&
              "cursor-pointer hover:bg-accent hover:text-accent-foreground"
          )}
          onClick={() =>
            log.domain && !log.hostSet && onDomainClick(log.domain)
          }
        >
          <div className="flex flex-row gap-2 items-center">
            {log.domain && <span>{log.domain}</span>}
            <div className="flex-1" />
            {log.domain && !log.hostSet && (
              <AddIcon
                className="h-4 w-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                onClick={(e) => {
                  e.stopPropagation();
                  onDomainClick(log.domain!);
                }}
              />
            )}
          </div>
        </td>
        <td className="text-muted-foreground font-mono text-xs border-b border-border py-2 px-4">
          <Tooltip>
            <TooltipTrigger asChild>
              <span>
                {log.deviceName ? (
                  <Badge>{log.deviceName}</Badge>
                ) : (
                  <span>{log.source}</span>
                )}
              </span>
            </TooltipTrigger>
            <TooltipContent>
              <p>{log.source}</p>
            </TooltipContent>
          </Tooltip>
        </td>
        <td className="text-foreground border-b border-border py-2 px-4">
          <div className="flex flex-row gap-2 items-center">
            <span
              className={cn(
                !log.ipSet &&
                  "cursor-pointer hover:bg-accent hover:text-accent-foreground"
              )}
              onClick={() =>
                log.destination && !log.ipSet && onIpClick(log.destination)
              }
            >
              {log.destination}
            </span>
            {asnName && <Badge variant="outline">{asnName}</Badge>}
            <div className="flex-1" />
            {!log.ipSet && (
              <AddIcon
                className="h-4 w-4 cursor-pointer text-muted-foreground hover:text-foreground transition-colors"
                onClick={(e) => {
                  e.stopPropagation();
                  onIpClick(log.destination!);
                }}
              />
            )}
          </div>
        </td>
      </tr>
    );
  },
  (prev, next) => prev.log.raw === next.log.raw
);

TableRowMemo.displayName = "TableRowMemo";

export const DomainsTable = ({
  data,
  sortColumn,
  sortDirection,
  onSort,
  onDomainClick,
  onIpClick,
  onScrollStateChange,
}: DomainsTableProps) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [scrollTop, setScrollTop] = useState(0);
  const [containerHeight, setContainerHeight] = useState(600);

  const startIndex = Math.max(0, Math.floor(scrollTop / ROW_HEIGHT) - OVERSCAN);
  const visibleCount = Math.ceil(containerHeight / ROW_HEIGHT) + OVERSCAN * 2;
  const endIndex = Math.min(data.length, startIndex + visibleCount);

  const visibleData = useMemo(
    () => data.slice(startIndex, endIndex),
    [data, startIndex, endIndex]
  );

  const handleScroll = useCallback(
    (e: React.UIEvent<HTMLDivElement>) => {
      const target = e.currentTarget;
      setScrollTop(target.scrollTop);

      const isAtBottom =
        target.scrollHeight - target.scrollTop - target.clientHeight < 50;
      onScrollStateChange(isAtBottom);
    },
    [onScrollStateChange]
  );

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerHeight(entry.contentRect.height);
      }
    });

    observer.observe(container);
    setContainerHeight(container.clientHeight);

    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    requestAnimationFrame(() => {
      const isAtBottom =
        container.scrollHeight - container.scrollTop - container.clientHeight <
        100;
      if (isAtBottom) {
        container.scrollTop = container.scrollHeight;
        setScrollTop(container.scrollTop);
      }
    });
  }, [data.length]);

  return (
    <div
      ref={containerRef}
      onScroll={handleScroll}
      className="flex-1 bg-background overflow-auto"
    >
      <table className="w-full border-collapse">
        <thead className="sticky top-0 z-1">
          <tr>
            <SortableTableCell
              label="Time"
              active={sortColumn === "timestamp"}
              direction={sortColumn === "timestamp" ? sortDirection : null}
              onSort={() => onSort("timestamp")}
            />
            <SortableTableCell
              label="Protocol"
              active={sortColumn === "protocol"}
              direction={sortColumn === "protocol" ? sortDirection : null}
              onSort={() => onSort("protocol")}
            />
            <SortableTableCell
              label="Set"
              active={sortColumn === "set"}
              direction={sortColumn === "set" ? sortDirection : null}
              onSort={() => onSort("set")}
            />
            <SortableTableCell
              label="Domain"
              active={sortColumn === "domain"}
              direction={sortColumn === "domain" ? sortDirection : null}
              onSort={() => onSort("domain")}
            />
            <SortableTableCell
              label="Source"
              active={sortColumn === "source"}
              direction={sortColumn === "source" ? sortDirection : null}
              onSort={() => onSort("source")}
            />
            <SortableTableCell
              label="Destination"
              active={sortColumn === "destination"}
              direction={sortColumn === "destination" ? sortDirection : null}
              onSort={() => onSort("destination")}
            />
          </tr>
        </thead>
        <tbody>
          {data.length === 0 ? (
            <tr>
              <td
                colSpan={6}
                className="text-center py-8 text-muted-foreground italic bg-background border-none"
              >
                Waiting for connections...
              </td>
            </tr>
          ) : (
            <>
              {startIndex > 0 && (
                <tr>
                  <td
                    colSpan={6}
                    style={{ height: startIndex * ROW_HEIGHT }}
                    className="p-0 border-none"
                  />
                </tr>
              )}

              {visibleData.map((log) => (
                <TableRowMemo
                  key={log.raw}
                  log={log}
                  onDomainClick={onDomainClick}
                  onIpClick={onIpClick}
                />
              ))}

              {endIndex < data.length && (
                <tr>
                  <td
                    colSpan={6}
                    style={{ height: (data.length - endIndex) * ROW_HEIGHT }}
                    className="p-0 border-none"
                  />
                </tr>
              )}
            </>
          )}
        </tbody>
      </table>
    </div>
  );
};
