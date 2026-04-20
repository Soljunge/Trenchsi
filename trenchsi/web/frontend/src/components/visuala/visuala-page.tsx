import {
  IconBrain,
  IconClockHour4,
  IconPlayerPause,
  IconPlugConnected,
  IconTerminal2,
  IconTrendingUp,
} from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import type { ComponentType } from "react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getAppConfig } from "@/api/channels"
import { PageHeader } from "@/components/page-header"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import { useChatModels } from "@/hooks/use-chat-models"
import { useGateway } from "@/hooks/use-gateway"
import { useGatewayLogs } from "@/hooks/use-gateway-logs"
import { cn } from "@/lib/utils"

import agentAvatarUrl from "../../../../../assets/agent-avatar.png"

type ActivityTone = "thinking" | "tool" | "io" | "idle"

interface ActivityItem {
  id: string
  label: string
  detail: string
  tone: ActivityTone
  Icon: ComponentType<{ className?: string }>
}

type TradingConfig = Record<string, unknown>

const MAX_EVENTS = 12

export function VisualaPage() {
  const { t } = useTranslation()
  const { state } = useGateway()
  const { logs } = useGatewayLogs()
  const { defaultModel } = useChatModels({ isConnected: state === "running" })
  const {
    data: appConfig,
    isLoading: isTradingConfigLoading,
    error: tradingConfigError,
  } = useQuery({
    queryKey: ["app-config"],
    queryFn: getAppConfig,
    staleTime: 30_000,
  })
  const [showSummary, setShowSummary] = useState(true)
  const [showFlow, setShowFlow] = useState(true)
  const [showTimeline, setShowTimeline] = useState(true)
  const [showThinking, setShowThinking] = useState(true)
  const [showTools, setShowTools] = useState(true)
  const [showMemory, setShowMemory] = useState(true)
  const [showTrading, setShowTrading] = useState(true)
  const [showAgentIcon, setShowAgentIcon] = useState(true)
  const [showCustomize, setShowCustomize] = useState(false)

  const recentLogs = logs.slice(-MAX_EVENTS).reverse()
  const visualEvents = recentLogs.map((line, index) =>
    createActivityItem(line, `${logs.length - index}`),
  )

  const metrics = summarizeLogs(logs, state)
  const providerLabel = describeProvider(defaultModel)
  const activeFlowStep = resolveActiveFlowStep(metrics)
  const tradingConfig = asRecord(asRecord(appConfig).trading) as TradingConfig
  const tradingStatusLabel = describeTradingStatus(tradingConfig, t)
  const tradingModeLabel = describeTradingMode(tradingConfig, t)

  useEffect(() => {
    const handleToggle = () => {
      setShowCustomize((prev) => !prev)
    }

    window.addEventListener("visuala:toggle-customize", handleToggle)
    return () => {
      window.removeEventListener("visuala:toggle-customize", handleToggle)
    }
  }, [])

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("navigation.visuala")} />

      <div className="flex-1 overflow-auto px-6 py-3">
        <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 pb-8">
          {showCustomize ? (
            <Card className="border-border/60 border" size="sm">
              <CardHeader>
                <CardTitle>
                  {t("pages.agent.visuala.customize_title")}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-3 md:grid-cols-3">
                  <ToggleRow
                    label={t("pages.agent.visuala.customize.summary")}
                    checked={showSummary}
                    onCheckedChange={setShowSummary}
                  />
                  <ToggleRow
                    label={t("pages.agent.visuala.customize.flow")}
                    checked={showFlow}
                    onCheckedChange={setShowFlow}
                  />
                  <ToggleRow
                    label={t("pages.agent.visuala.customize.timeline")}
                    checked={showTimeline}
                    onCheckedChange={setShowTimeline}
                  />
                  <ToggleRow
                    label={t("pages.agent.visuala.cards.thinking")}
                    checked={showThinking}
                    onCheckedChange={setShowThinking}
                  />
                  <ToggleRow
                    label={t("pages.agent.visuala.cards.tools")}
                    checked={showTools}
                    onCheckedChange={setShowTools}
                  />
                  <ToggleRow
                    label={t("pages.agent.visuala.cards.memory")}
                    checked={showMemory}
                    onCheckedChange={setShowMemory}
                  />
                  <ToggleRow
                    label={t("pages.agent.visuala.customize.trading")}
                    checked={showTrading}
                    onCheckedChange={setShowTrading}
                  />
                  <ToggleRow
                    label={t("pages.agent.visuala.agent_icon.title")}
                    checked={showAgentIcon}
                    onCheckedChange={setShowAgentIcon}
                  />
                </div>
              </CardContent>
            </Card>
          ) : null}

          {showSummary ? (
            <Card className="border-border/60 border" size="sm">
              <CardHeader className="flex flex-row items-start justify-between gap-4">
                <div>
                  <CardTitle>{t("pages.agent.visuala.hero_title")}</CardTitle>
                </div>
                {showAgentIcon ? (
                  <AgentIconDisplay
                    title={t("pages.agent.visuala.agent_icon.title")}
                    imageAlt={t("pages.agent.visuala.agent_icon.alt")}
                  />
                ) : null}
              </CardHeader>
              <CardContent>
                <div className="grid gap-3 md:grid-cols-5">
                  <SimpleBox
                    label={t("pages.agent.visuala.metrics.gateway")}
                    value={t(`pages.agent.visuala.gateway_state.${state}`)}
                  />
                  <SimpleBox
                    label={t("pages.agent.visuala.metrics.phase")}
                    value={t(`pages.agent.visuala.phase.${metrics.phase}`)}
                  />
                  <SimpleBox
                    label={t("pages.agent.visuala.metrics.tokens")}
                    value={formatTokenUsage(
                      metrics.tokensUsed,
                      t("pages.agent.visuala.metrics.tokens_unknown"),
                    )}
                  />
                  {showThinking ? (
                    <SimpleBox
                      label={t("pages.agent.visuala.cards.thinking")}
                      value={String(metrics.thinkingCount)}
                    />
                  ) : null}
                  {showTools ? (
                    <SimpleBox
                      label={t("pages.agent.visuala.cards.tools")}
                      value={String(metrics.toolCount)}
                    />
                  ) : null}
                  {showMemory ? (
                    <SimpleBox
                      label={t("pages.agent.visuala.cards.memory")}
                      value={t(
                        `pages.agent.visuala.memory_state.${metrics.memoryState}`,
                      )}
                    />
                  ) : null}
                </div>
              </CardContent>
            </Card>
          ) : null}

          {showTrading ? (
            <Card className="border-border/60 border" size="sm">
              <CardHeader className="flex flex-row items-start justify-between gap-4">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <IconTrendingUp className="size-4 text-emerald-600" />
                    {t("pages.agent.visuala.trading.title")}
                  </CardTitle>
                  <CardDescription>
                    {t("pages.agent.visuala.trading.description")}
                  </CardDescription>
                </div>
                <div className="flex flex-wrap justify-end gap-2 text-[11px] tracking-[0.18em] uppercase">
                  <TradingBadge tone={tradingStatusTone(tradingConfig)}>
                    {tradingStatusLabel}
                  </TradingBadge>
                  <TradingBadge
                    tone={asBool(tradingConfig.dry_run) ? "warning" : "success"}
                  >
                    {tradingModeLabel}
                  </TradingBadge>
                </div>
              </CardHeader>
              <CardContent>
                {tradingConfigError ? (
                  <EmptyState text={t("pages.agent.visuala.trading.error")} />
                ) : isTradingConfigLoading ? (
                  <EmptyState text={t("labels.loading")} />
                ) : (
                  <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                    <SimpleBox
                      label={t("pages.agent.visuala.trading.fields.strategy")}
                      value={formatText(
                        tradingConfig.strategy,
                        t("pages.agent.visuala.trading.fallbacks.strategy"),
                      )}
                    />
                    <SimpleBox
                      label={t("pages.agent.visuala.trading.fields.network")}
                      value={formatText(
                        tradingConfig.network,
                        t("pages.agent.visuala.trading.fallbacks.network"),
                      )}
                    />
                    <SimpleBox
                      label={t("pages.agent.visuala.trading.fields.position")}
                      value={formatSol(
                        tradingConfig.max_position_sol,
                        t("pages.agent.visuala.trading.fallbacks.position"),
                      )}
                    />
                    <SimpleBox
                      label={t("pages.agent.visuala.trading.fields.slippage")}
                      value={formatBps(
                        tradingConfig.default_slippage_bps,
                        t("pages.agent.visuala.trading.fallbacks.slippage"),
                      )}
                    />
                    <SimpleBox
                      label={t("pages.agent.visuala.trading.fields.risk")}
                      value={formatRiskSummary(tradingConfig, t)}
                    />
                    <SimpleBox
                      label={t("pages.agent.visuala.trading.fields.filters")}
                      value={formatFilterSummary(tradingConfig, t)}
                    />
                  </div>
                )}
              </CardContent>
            </Card>
          ) : null}

          {(showSummary || showTrading) && (showFlow || showTimeline) ? (
            <div className="flex justify-center">
              <div className="bg-border h-8 w-px" />
            </div>
          ) : null}

          {showFlow || showTimeline ? (
            <div className="flex flex-col gap-6">
              {showFlow ? (
                <Card className="border-border/60 border" size="sm">
                  <CardHeader>
                    <CardTitle>{t("pages.agent.visuala.flow.title")}</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid gap-3 lg:grid-cols-[1fr_56px_1fr_56px_1fr] lg:items-center">
                      <FlowStage
                        label={t("pages.agent.visuala.flow.step_focus")}
                        title={t("pages.agent.visuala.now_title")}
                        value={t(`pages.agent.visuala.phase.${metrics.phase}`)}
                        tone={metrics.phaseTone}
                        active={activeFlowStep === "focus"}
                      />
                      <FlowConnector />
                      <FlowStage
                        label={t("pages.agent.visuala.flow.step_agent")}
                        title={t("pages.agent.visuala.flow.new_agent")}
                        value={providerLabel}
                        tone="tool"
                        active={activeFlowStep === "agent"}
                      />
                      <FlowConnector />
                      <FlowStage
                        label={t("pages.agent.visuala.flow.step_memory")}
                        title={t("pages.agent.visuala.flow.memory_target")}
                        value={metrics.memoryTarget}
                        tone="io"
                        active={activeFlowStep === "memory"}
                      />
                    </div>

                    <div className="grid gap-3 md:grid-cols-2">
                      <SimpleFlowBox
                        title={t("pages.agent.visuala.cards.io")}
                        value={String(metrics.ioCount)}
                        tone="io"
                      />
                      <SimpleFlowBox
                        title={t("pages.agent.visuala.flow.activity")}
                        value={metrics.summary}
                        tone="idle"
                      />
                    </div>
                  </CardContent>
                </Card>
              ) : null}

              {showTimeline ? (
                <Card className="border-border/60 border" size="sm">
                  <CardHeader>
                    <CardTitle>
                      {t("pages.agent.visuala.timeline_title")}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    {visualEvents.length > 0 ? (
                      <div className="space-y-0">
                        {visualEvents.map((event, index) => (
                          <div key={event.id} className="relative pl-8">
                            {index < visualEvents.length - 1 ? (
                              <div className="bg-border absolute top-8 left-3 h-[calc(100%+0.5rem)] w-px" />
                            ) : null}
                            <div className="bg-background border-border absolute top-3 left-0 flex size-6 items-center justify-center border">
                              <event.Icon className="size-3.5" />
                            </div>
                            <div
                              className={cn(
                                "mb-3 border px-3 py-3",
                                boxToneClasses[event.tone],
                              )}
                            >
                              <div className="text-sm font-medium">
                                {event.label}
                              </div>
                              <div className="mt-1 font-mono text-xs break-words opacity-80">
                                {event.detail}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <EmptyState text={t("pages.agent.visuala.empty")} />
                    )}
                  </CardContent>
                </Card>
              ) : null}
            </div>
          ) : (
            <EmptyState
              text={t("pages.agent.visuala.customize.nothing_selected")}
            />
          )}
        </div>
      </div>
    </div>
  )
}

function ToggleRow({
  label,
  checked,
  onCheckedChange,
}: {
  label: string
  checked: boolean
  onCheckedChange: (checked: boolean) => void
}) {
  return (
    <label className="border-border flex items-center justify-between border px-3 py-3 text-sm">
      <span>{label}</span>
      <Switch checked={checked} onCheckedChange={onCheckedChange} size="sm" />
    </label>
  )
}

function AgentIconDisplay({
  title,
  imageAlt,
}: {
  title: string
  imageAlt: string
}) {
  const [imageFailed, setImageFailed] = useState(false)

  return (
    <div className="flex shrink-0 items-center gap-3 rounded-2xl border border-orange-200/70 bg-orange-50/60 px-3 py-2">
      <div className="flex size-14 shrink-0 items-center justify-center overflow-hidden rounded-2xl border border-orange-200 bg-gradient-to-br from-amber-100 via-orange-50 to-white shadow-[0_10px_30px_-18px_rgba(194,65,12,0.7)]">
        {imageFailed ? (
          <span className="text-2xl">🪖</span>
        ) : (
          <img
            src={agentAvatarUrl}
            alt={imageAlt}
            className="size-full object-cover"
            onError={() => setImageFailed(true)}
          />
        )}
      </div>
      <div className="hidden min-w-0 sm:block">
        <div className="text-[11px] tracking-[0.18em] text-orange-700 uppercase">
          {title}
        </div>
        <div className="mt-1 text-sm font-medium">Trenchsi</div>
      </div>
    </div>
  )
}

function SimpleBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="border-border border px-3 py-3">
      <div className="text-muted-foreground text-[11px] tracking-[0.18em] uppercase">
        {label}
      </div>
      <div className="mt-2 text-base font-medium">{value}</div>
    </div>
  )
}

function TradingBadge({
  tone,
  children,
}: {
  tone: "success" | "warning" | "neutral"
  children: string
}) {
  const toneClasses: Record<"success" | "warning" | "neutral", string> = {
    success: "border-emerald-200 bg-emerald-50 text-emerald-900",
    warning: "border-amber-200 bg-amber-50 text-amber-900",
    neutral: "border-border bg-background text-foreground",
  }

  return (
    <span className={cn("rounded-full border px-2.5 py-1", toneClasses[tone])}>
      {children}
    </span>
  )
}

function SimpleFlowBox({
  title,
  value,
  tone,
}: {
  title: string
  value: string
  tone: ActivityTone
}) {
  return (
    <div className={cn("border px-4 py-4 text-center", boxToneClasses[tone])}>
      <div className="text-xs tracking-[0.18em] uppercase opacity-70">
        {title}
      </div>
      <div className="mt-2 text-sm font-medium">{value}</div>
    </div>
  )
}

function FlowStage({
  label,
  title,
  value,
  tone,
  active,
}: {
  label: string
  title: string
  value: string
  tone: ActivityTone
  active: boolean
}) {
  return (
    <div className="space-y-2">
      <div className="text-muted-foreground flex items-center justify-center gap-2 text-center text-[11px] font-medium tracking-[0.18em] uppercase">
        {active ? (
          <span className="size-3 animate-spin rounded-full border-2 border-current border-t-transparent" />
        ) : (
          <span className="bg-border size-2 rounded-full" />
        )}
        {label}
      </div>
      <SimpleFlowBox title={title} value={value} tone={tone} />
    </div>
  )
}

function FlowConnector() {
  return (
    <div className="hidden items-center justify-center lg:flex">
      <div className="bg-border h-px w-full" />
      <div className="border-border bg-background -ml-2 size-2 rotate-45 border-r border-b" />
    </div>
  )
}

function EmptyState({ text }: { text: string }) {
  return (
    <div className="text-muted-foreground rounded-xl border border-dashed px-4 py-8 text-center text-sm">
      {text}
    </div>
  )
}

const boxToneClasses: Record<ActivityTone, string> = {
  thinking: "border-violet-300 bg-violet-50 text-violet-900",
  tool: "border-emerald-300 bg-emerald-50 text-emerald-900",
  io: "border-sky-300 bg-sky-50 text-sky-900",
  idle: "border-border bg-background text-foreground",
}

function createActivityItem(line: string, id: string): ActivityItem {
  const lower = line.toLowerCase()

  if (
    lower.includes("tool") ||
    lower.includes("exec") ||
    lower.includes("command") ||
    lower.includes("mcp")
  ) {
    return {
      id,
      label: "Tool execution",
      detail: line,
      tone: "tool",
      Icon: IconTerminal2,
    }
  }

  if (
    lower.includes("thinking") ||
    lower.includes("reason") ||
    lower.includes("model") ||
    lower.includes("assistant")
  ) {
    return {
      id,
      label: "Reasoning",
      detail: line,
      tone: "thinking",
      Icon: IconBrain,
    }
  }

  if (
    lower.includes("session") ||
    lower.includes("connect") ||
    lower.includes("channel") ||
    lower.includes("message")
  ) {
    return {
      id,
      label: "I/O activity",
      detail: line,
      tone: "io",
      Icon: IconPlugConnected,
    }
  }

  return {
    id,
    label: "Background activity",
    detail: line,
    tone: "idle",
    Icon: IconClockHour4,
  }
}

function asRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>
  }
  return {}
}

function asBool(value: unknown): boolean {
  return value === true
}

function asNumber(value: unknown): number | undefined {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined
}

function asString(value: unknown): string {
  return typeof value === "string" ? value.trim() : ""
}

function formatText(value: unknown, fallback: string): string {
  const text = asString(value)
  return text !== "" ? text : fallback
}

function formatSol(value: unknown, fallback: string): string {
  const number = asNumber(value)
  if (number === undefined) return fallback
  return `${number.toFixed(number >= 1 ? 2 : 4)} SOL`
}

function formatBps(value: unknown, fallback: string): string {
  const number = asNumber(value)
  if (number === undefined) return fallback
  return `${Math.round(number)} bps`
}

function formatPercent(value: unknown, fallback: string): string {
  const number = asNumber(value)
  if (number === undefined) return fallback
  return `${number.toFixed(number >= 10 ? 0 : 1)}%`
}

function formatCurrency(value: unknown, fallback: string): string {
  const number = asNumber(value)
  if (number === undefined) return fallback
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 0,
  }).format(number)
}

function formatDurationSeconds(value: unknown, fallback: string): string {
  const number = asNumber(value)
  if (number === undefined) return fallback
  if (number % 60 === 0) {
    return `${number / 60}m`
  }
  return `${number}s`
}

function formatRiskSummary(
  tradingConfig: TradingConfig,
  t: (key: string) => string,
): string {
  return [
    `${t("pages.agent.visuala.trading.fields.take_profit")}: ${formatPercent(tradingConfig.take_profit_pct, t("pages.agent.visuala.trading.fallbacks.take_profit"))}`,
    `${t("pages.agent.visuala.trading.fields.stop_loss")}: ${formatPercent(tradingConfig.stop_loss_pct, t("pages.agent.visuala.trading.fallbacks.stop_loss"))}`,
    `${t("pages.agent.visuala.trading.fields.trailing_stop")}: ${formatPercent(tradingConfig.trailing_stop_pct, t("pages.agent.visuala.trading.fallbacks.trailing_stop"))}`,
  ].join(" · ")
}

function formatFilterSummary(
  tradingConfig: TradingConfig,
  t: (key: string) => string,
): string {
  return [
    `${t("pages.agent.visuala.trading.fields.liquidity")}: ${formatCurrency(tradingConfig.min_liquidity_usd, t("pages.agent.visuala.trading.fallbacks.liquidity"))}`,
    `${t("pages.agent.visuala.trading.fields.volume")}: ${formatCurrency(tradingConfig.min_volume_usd, t("pages.agent.visuala.trading.fallbacks.volume"))}`,
    `${t("pages.agent.visuala.trading.fields.cooldown")}: ${formatDurationSeconds(tradingConfig.trade_cooldown_seconds, t("pages.agent.visuala.trading.fallbacks.cooldown"))}`,
  ].join(" · ")
}

function describeTradingStatus(
  tradingConfig: TradingConfig,
  t: (key: string) => string,
): string {
  const parts = [
    asBool(tradingConfig.enabled)
      ? t("pages.agent.visuala.trading.status.enabled")
      : t("pages.agent.visuala.trading.status.disabled"),
    asBool(tradingConfig.enable_paper_trading)
      ? t("pages.agent.visuala.trading.status.paper")
      : t("pages.agent.visuala.trading.status.live"),
  ]

  if (asBool(tradingConfig.dry_run)) {
    parts.push(t("pages.agent.visuala.trading.status.dry_run"))
  }
  if (asBool(tradingConfig.emergency_halt)) {
    parts.push(t("pages.agent.visuala.trading.status.halted"))
  }

  return parts.join(" · ")
}

function describeTradingMode(
  tradingConfig: TradingConfig,
  t: (key: string) => string,
): string {
  if (asBool(tradingConfig.emergency_halt)) {
    return t("pages.agent.visuala.trading.mode.halted")
  }

  const bits = []
  if (asBool(tradingConfig.enable_paper_trading)) {
    bits.push(t("pages.agent.visuala.trading.mode.paper"))
  }
  if (asBool(tradingConfig.dry_run)) {
    bits.push(t("pages.agent.visuala.trading.mode.dry_run"))
  }
  if (bits.length === 0) {
    bits.push(t("pages.agent.visuala.trading.mode.live"))
  }
  return bits.join(" · ")
}

function tradingStatusTone(
  tradingConfig: TradingConfig,
): "success" | "warning" | "neutral" {
  if (asBool(tradingConfig.emergency_halt)) {
    return "warning"
  }
  if (asBool(tradingConfig.enabled)) {
    return "success"
  }
  return "neutral"
}

function summarizeLogs(logs: string[], gatewayState: string) {
  const thinkingCount = countMatches(logs, [
    "thinking",
    "reason",
    "assistant",
    "model",
  ])
  const toolCount = countMatches(logs, ["tool", "exec", "command", "mcp"])
  const ioCount = countMatches(logs, [
    "message",
    "session",
    "channel",
    "connect",
  ])
  const memoryCount = countMatches(logs, [
    "memory",
    "memories",
    "store",
    "stored",
    "save",
    "saved",
    "recall",
  ])
  const tokensUsed = summarizeTokenUsage(logs)

  const latest = logs[logs.length - 1]
  const latestEvent = latest
    ? createActivityItem(latest, "latest")
    : {
        tone: "idle" as ActivityTone,
        label: "Idle",
        detail: "",
        Icon: IconPlayerPause,
      }

  let phase: "booting" | "thinking" | "acting" | "waiting" = "waiting"
  let phaseTone: ActivityTone = "idle"

  if (gatewayState === "starting" || gatewayState === "restarting") {
    phase = "booting"
    phaseTone = "io"
  } else if (latestEvent.tone === "thinking") {
    phase = "thinking"
    phaseTone = "thinking"
  } else if (latestEvent.tone === "tool") {
    phase = "acting"
    phaseTone = "tool"
  } else if (latestEvent.tone === "io") {
    phase = "booting"
    phaseTone = "io"
  }

  const summary = latest
    ? latest
    : "No live gateway activity has been captured yet."
  const memoryTarget = describeMemoryTarget(logs)

  const memoryState: "active" | "saved" | "idle" =
    latest &&
    containsAny(latest, [
      "memory",
      "memories",
      "store",
      "stored",
      "save",
      "saved",
      "recall",
    ])
      ? "active"
      : memoryCount > 0
        ? "saved"
        : "idle"

  return {
    thinkingCount,
    toolCount,
    ioCount,
    memoryCount,
    tokensUsed,
    memoryState,
    memoryTarget,
    phase,
    phaseTone,
    summary,
  }
}

function summarizeTokenUsage(logs: string[]) {
  let total = 0
  let found = false

  for (const line of logs) {
    const normalized = line.toLowerCase()
    const totalTokens = extractTokenValue(normalized, [
      /total_tokens["=: ]+(\d+)/g,
      /total tokens["=: ]+(\d+)/g,
      /tokens used["=: ]+(\d+)/g,
      /token usage["=: ]+(\d+)/g,
    ])

    if (totalTokens !== null) {
      total += totalTokens
      found = true
      continue
    }

    const inputTokens = extractTokenValue(normalized, [
      /prompt_tokens["=: ]+(\d+)/g,
      /input_tokens["=: ]+(\d+)/g,
      /prompt tokens["=: ]+(\d+)/g,
      /input tokens["=: ]+(\d+)/g,
    ])
    const outputTokens = extractTokenValue(normalized, [
      /completion_tokens["=: ]+(\d+)/g,
      /output_tokens["=: ]+(\d+)/g,
      /completion tokens["=: ]+(\d+)/g,
      /output tokens["=: ]+(\d+)/g,
    ])

    if (inputTokens !== null || outputTokens !== null) {
      total += (inputTokens ?? 0) + (outputTokens ?? 0)
      found = true
    }
  }

  return found ? total : null
}

function extractTokenValue(line: string, patterns: RegExp[]) {
  for (const pattern of patterns) {
    const match = pattern.exec(line)
    if (!match) {
      continue
    }

    const value = Number.parseInt(match[1], 10)
    if (Number.isFinite(value)) {
      return value
    }
  }

  return null
}

function formatTokenUsage(value: number | null, fallback: string) {
  if (value === null) {
    return fallback
  }

  return new Intl.NumberFormat().format(value)
}

function countMatches(lines: string[], patterns: string[]) {
  return lines.reduce((count, line) => {
    return containsAny(line, patterns) ? count + 1 : count
  }, 0)
}

function containsAny(line: string, patterns: string[]) {
  const lower = line.toLowerCase()
  return patterns.some((pattern) => lower.includes(pattern))
}

function describeProvider(
  model: {
    model_name?: string
    model?: string
    auth_method?: string
    api_base?: string
  } | null,
) {
  if (!model) {
    return "No provider"
  }

  const apiBase = model.api_base?.toLowerCase() ?? ""
  const modelId = model.model?.toLowerCase() ?? ""

  if (model.auth_method === "oauth") {
    return "OAuth provider"
  }
  if (model.auth_method === "local" || apiBase.includes("localhost")) {
    return "Local provider"
  }
  if (apiBase.includes("openrouter") || modelId.includes("openrouter")) {
    return "OpenRouter"
  }
  if (apiBase.includes("openai") || modelId.includes("openai")) {
    return "OpenAI"
  }
  if (apiBase.includes("anthropic") || modelId.includes("anthropic")) {
    return "Anthropic"
  }

  return model.model_name || model.model || "Configured provider"
}

function describeMemoryTarget(logs: string[]) {
  const joined = logs.slice(-20).join("\n").toLowerCase()

  if (joined.includes("workspace")) {
    return "Workspace memory"
  }
  if (joined.includes("session")) {
    return "Session store"
  }
  if (joined.includes("jsonl")) {
    return "JSONL memory"
  }
  if (joined.includes("memory") || joined.includes("saved")) {
    return "Memory store"
  }

  return "Waiting for memory write"
}

function resolveActiveFlowStep(metrics: {
  phase: "booting" | "thinking" | "acting" | "waiting"
  memoryState: "active" | "saved" | "idle"
}) {
  if (metrics.memoryState === "active") {
    return "memory"
  }
  if (metrics.phase === "acting" || metrics.phase === "booting") {
    return "agent"
  }
  return "focus"
}
