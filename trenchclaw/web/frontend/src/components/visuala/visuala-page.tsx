import {
  IconBrain,
  IconClockHour4,
  IconPlayerPause,
  IconPlugConnected,
  IconTerminal2,
} from "@tabler/icons-react"
import type { ComponentType } from "react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import agentAvatarUrl from "../../../../../assets/agent-avatar.png"

import { PageHeader } from "@/components/page-header"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import { useChatModels } from "@/hooks/use-chat-models"
import { useGateway } from "@/hooks/use-gateway"
import { useGatewayLogs } from "@/hooks/use-gateway-logs"
import { cn } from "@/lib/utils"

type ActivityTone = "thinking" | "tool" | "io" | "idle"

interface ActivityItem {
  id: string
  label: string
  detail: string
  tone: ActivityTone
  Icon: ComponentType<{ className?: string }>
}

const MAX_EVENTS = 12

export function VisualaPage() {
  const { t } = useTranslation()
  const { state } = useGateway()
  const { logs } = useGatewayLogs()
  const { defaultModel } = useChatModels({ isConnected: state === "running" })
  const [showSummary, setShowSummary] = useState(true)
  const [showFlow, setShowFlow] = useState(true)
  const [showTimeline, setShowTimeline] = useState(true)
  const [showThinking, setShowThinking] = useState(true)
  const [showTools, setShowTools] = useState(true)
  const [showMemory, setShowMemory] = useState(true)
  const [showAgentIcon, setShowAgentIcon] = useState(true)
  const [showCustomize, setShowCustomize] = useState(false)

  const recentLogs = logs.slice(-MAX_EVENTS).reverse()
  const visualEvents = recentLogs.map((line, index) =>
    createActivityItem(line, `${logs.length - index}`),
  )

  const metrics = summarizeLogs(logs, state)
  const providerLabel = describeProvider(defaultModel)
  const activeFlowStep = resolveActiveFlowStep(metrics)

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
            <Card className="border border-border/60" size="sm">
              <CardHeader>
                <CardTitle>{t("pages.agent.visuala.customize_title")}</CardTitle>
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
                    label={t("pages.agent.visuala.agent_icon.title")}
                    checked={showAgentIcon}
                    onCheckedChange={setShowAgentIcon}
                  />
                </div>
              </CardContent>
            </Card>
          ) : null}

          {showSummary ? (
          <Card className="border border-border/60" size="sm">
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

          {showSummary && (showFlow || showTimeline) ? (
            <div className="flex justify-center">
              <div className="bg-border h-8 w-px" />
            </div>
          ) : null}

          {showFlow || showTimeline ? (
            <div className="flex flex-col gap-6">
              {showFlow ? (
                <Card className="border border-border/60" size="sm">
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
                <Card className="border border-border/60" size="sm">
                  <CardHeader>
                    <CardTitle>{t("pages.agent.visuala.timeline_title")}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {visualEvents.length > 0 ? (
                      <div className="space-y-0">
                        {visualEvents.map((event, index) => (
                          <div key={event.id} className="relative pl-8">
                            {index < visualEvents.length - 1 ? (
                              <div className="bg-border absolute top-8 left-3 h-[calc(100%+0.5rem)] w-px" />
                            ) : null}
                            <div className="bg-background absolute top-3 left-0 flex size-6 items-center justify-center border border-border">
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
    <label className="flex items-center justify-between border border-border px-3 py-3 text-sm">
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
      <div className="from-amber-100 via-orange-50 to-white flex size-14 shrink-0 items-center justify-center overflow-hidden rounded-2xl border border-orange-200 bg-gradient-to-br shadow-[0_10px_30px_-18px_rgba(194,65,12,0.7)]">
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
        <div className="text-[11px] uppercase tracking-[0.18em] text-orange-700">
          {title}
        </div>
        <div className="mt-1 text-sm font-medium">TrenchClaw</div>
      </div>
    </div>
  )
}

function SimpleBox({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-border px-3 py-3">
      <div className="text-muted-foreground text-[11px] uppercase tracking-[0.18em]">
        {label}
      </div>
      <div className="mt-2 text-base font-medium">{value}</div>
    </div>
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
    <div
      className={cn(
        "border px-4 py-4 text-center",
        boxToneClasses[tone],
      )}
    >
      <div className="text-xs uppercase tracking-[0.18em] opacity-70">
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
      <div className="text-muted-foreground flex items-center justify-center gap-2 text-center text-[11px] font-medium uppercase tracking-[0.18em]">
        {active ? (
          <span className="size-3 rounded-full border-2 border-current border-t-transparent animate-spin" />
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
    memoryState,
    memoryTarget,
    phase,
    phaseTone,
    summary,
  }
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
