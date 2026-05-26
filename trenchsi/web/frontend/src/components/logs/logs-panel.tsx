import { type RefObject, useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

import { AnsiLogLine } from "@/components/logs/ansi-log-line"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  type WebconsoleSettings,
  WEBCONSOLE_COLOR_SCHEMES,
} from "@/hooks/use-webconsole-settings"

type LogsPanelProps = {
  logs: string[]
  wrapColumns: number
  contentRef: RefObject<HTMLDivElement | null>
  measureRef: RefObject<HTMLSpanElement | null>
  settings: WebconsoleSettings
}

export function LogsPanel({
  logs,
  wrapColumns,
  contentRef,
  measureRef,
  settings,
}: LogsPanelProps) {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const colorScheme = WEBCONSOLE_COLOR_SCHEMES[settings.colorScheme]

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollIntoView({ behavior: "smooth" })
    }
  }, [logs])

  return (
    <div
      className="relative flex-1 overflow-hidden rounded-lg border"
      style={{
        backgroundColor: colorScheme.background,
        borderColor: colorScheme.border,
        color: colorScheme.foreground,
      }}
    >
      <ScrollArea className="h-full">
        <div
          ref={contentRef}
          className="relative p-4 font-mono leading-relaxed"
          style={{ fontSize: settings.fontSize }}
        >
          <span
            ref={measureRef}
            aria-hidden
            className="pointer-events-none invisible absolute font-mono"
            style={{ fontSize: settings.fontSize }}
          >
            0
          </span>
          {logs.length === 0 ? (
            <div className="italic" style={{ color: colorScheme.empty }}>
              {t("pages.logs.empty")}
            </div>
          ) : (
            logs.map((log, index) => (
              <AnsiLogLine key={index} line={log} wrapColumns={wrapColumns} />
            ))
          )}
          <div ref={scrollRef} />
        </div>
      </ScrollArea>
    </div>
  )
}
