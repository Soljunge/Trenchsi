import {
  IconBrain,
  IconFileInfo,
  IconSparkles,
} from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import ReactMarkdown from "react-markdown"
import rehypeRaw from "rehype-raw"
import rehypeSanitize from "rehype-sanitize"
import remarkGfm from "remark-gfm"

import {
  type SkillSupportItem,
  getSkill,
  getSkills,
} from "@/api/skills"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"

export function LearnedSkillsPage() {
  const { t } = useTranslation()
  const [selectedSkill, setSelectedSkill] = useState<SkillSupportItem | null>(
    null,
  )

  const { data, isLoading, error } = useQuery({
    queryKey: ["skills"],
    queryFn: getSkills,
  })

  const learnedSkills = useMemo(
    () =>
      [...(data?.skills ?? [])]
        .filter((skill) => skill.learned)
        .sort(
          (left, right) =>
            (right.origin?.installed_at ?? 0) - (left.origin?.installed_at ?? 0),
        ),
    [data],
  )

  const {
    data: selectedSkillDetail,
    isLoading: isSkillDetailLoading,
    error: skillDetailError,
  } = useQuery({
    queryKey: ["skills", selectedSkill?.name],
    queryFn: () => getSkill(selectedSkill!.name),
    enabled: selectedSkill !== null,
  })

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("navigation.learned")} />

      <div className="flex-1 overflow-auto px-6 py-3">
        <div className="w-full max-w-6xl space-y-6">
          {isLoading ? (
            <div className="text-muted-foreground py-6 text-sm">
              {t("labels.loading")}
            </div>
          ) : error ? (
            <div className="text-destructive py-6 text-sm">
              {t("pages.agent.load_error")}
            </div>
          ) : (
            <section className="space-y-5">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <p className="text-muted-foreground max-w-3xl text-sm">
                  {t("pages.agent.learned.description")}
                </p>
                <div className="rounded-full border border-orange-200 bg-orange-50 px-3 py-1 text-xs font-medium text-orange-800">
                  {t("pages.agent.learned.count", { count: learnedSkills.length })}
                </div>
              </div>

              {learnedSkills.length ? (
                <div className="grid gap-4 lg:grid-cols-2">
                  {learnedSkills.map((skill) => (
                    <Card
                      key={`${skill.source}:${skill.name}`}
                      className="border-border/60 gap-4 bg-white/80"
                      size="sm"
                    >
                      <CardHeader>
                        <div className="flex items-start justify-between gap-3">
                          <div className="space-y-3">
                            <div className="flex flex-wrap items-center gap-2">
                              <CardTitle className="font-semibold">
                                {skill.name}
                              </CardTitle>
                              <OriginChip skill={skill} />
                            </div>
                            <CardDescription>
                              {skill.description ||
                                t("pages.agent.learned.no_description")}
                            </CardDescription>
                          </div>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            className="text-muted-foreground hover:text-foreground"
                            onClick={() => setSelectedSkill(skill)}
                            title={t("pages.agent.learned.view")}
                          >
                            <IconFileInfo className="size-4" />
                          </Button>
                        </div>
                      </CardHeader>
                      <CardContent className="space-y-3">
                        <div className="flex flex-wrap gap-2 text-xs">
                          <MetaPill label={t("pages.agent.learned.source")} value={skill.source} />
                          <MetaPill
                            label={t("pages.agent.learned.learned_via")}
                            value={skill.learned_via || t("pages.agent.learned.manual")}
                          />
                          {skill.origin?.registry ? (
                            <MetaPill
                              label={t("pages.agent.learned.registry")}
                              value={skill.origin.registry}
                            />
                          ) : null}
                        </div>

                        {skill.origin?.installed_at ? (
                          <div className="text-muted-foreground text-xs">
                            {t("pages.agent.learned.installed_at", {
                              value: formatInstalledAt(skill.origin.installed_at),
                            })}
                          </div>
                        ) : null}

                        <div>
                          <div className="text-muted-foreground text-[11px] tracking-[0.18em] uppercase">
                            {t("pages.agent.learned.path")}
                          </div>
                          <div className="bg-muted/60 mt-2 overflow-x-auto rounded-lg px-3 py-2 font-mono text-xs leading-relaxed">
                            {skill.path}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : (
                <Card className="border-dashed">
                  <CardContent className="text-muted-foreground py-10 text-center text-sm">
                    {t("pages.agent.learned.empty")}
                  </CardContent>
                </Card>
              )}
            </section>
          )}
        </div>
      </div>

      <Sheet
        open={selectedSkill !== null}
        onOpenChange={(open) => {
          if (!open) setSelectedSkill(null)
        }}
      >
        <SheetContent
          side="right"
          className="w-full gap-0 p-0 data-[side=right]:!w-full data-[side=right]:sm:!w-[560px] data-[side=right]:sm:!max-w-[560px]"
        >
          <SheetHeader className="border-b px-6 py-5">
            <SheetTitle>
              {selectedSkill?.name || t("pages.agent.learned.viewer_title")}
            </SheetTitle>
            <SheetDescription>
              {selectedSkill?.description ||
                t("pages.agent.learned.viewer_description")}
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 overflow-auto px-6 py-5">
            {isSkillDetailLoading ? (
              <div className="text-muted-foreground text-sm">
                {t("pages.agent.learned.loading_detail")}
              </div>
            ) : skillDetailError ? (
              <div className="text-destructive text-sm">
                {t("pages.agent.learned.load_detail_error")}
              </div>
            ) : selectedSkillDetail ? (
              <div className="space-y-5">
                <CommandHints content={selectedSkillDetail.content} />

                <div className="prose prose-sm dark:prose-invert prose-pre:rounded-lg prose-pre:border prose-pre:bg-zinc-950 prose-pre:p-3 max-w-none">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    rehypePlugins={[rehypeRaw, rehypeSanitize]}
                  >
                    {selectedSkillDetail.content}
                  </ReactMarkdown>
                </div>
              </div>
            ) : null}
          </div>
        </SheetContent>
      </Sheet>
    </div>
  )
}

function OriginChip({ skill }: { skill: SkillSupportItem }) {
  if (skill.learned_via === "registry") {
    return (
      <span className="inline-flex items-center gap-1 rounded-full border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-800">
        <IconSparkles className="size-3" />
        registry
      </span>
    )
  }

  return (
    <span className="inline-flex items-center gap-1 rounded-full border border-orange-200 bg-orange-50 px-2 py-0.5 text-[11px] font-medium text-orange-800">
      <IconBrain className="size-3" />
      learned
    </span>
  )
}

function MetaPill({ label, value }: { label: string; value: string }) {
  return (
    <span className="inline-flex rounded-full border border-border bg-background px-2.5 py-1">
      <span className="text-muted-foreground mr-1">{label}:</span>
      <span className="font-medium">{value}</span>
    </span>
  )
}

function CommandHints({ content }: { content: string }) {
  const { t } = useTranslation()
  const commands = extractCommandHints(content)

  if (commands.length === 0) {
    return null
  }

  return (
    <div className="rounded-xl border border-orange-200 bg-orange-50/70 px-4 py-4">
      <div className="text-sm font-semibold text-orange-900">
        {t("pages.agent.learned.command_hints_title")}
      </div>
      <div className="text-muted-foreground mt-1 text-sm">
        {t("pages.agent.learned.command_hints_description")}
      </div>
      <div className="mt-3 space-y-2">
        {commands.map((command) => (
          <div
            key={command}
            className="overflow-x-auto rounded-lg border border-orange-200 bg-white px-3 py-2 font-mono text-xs text-orange-950"
          >
            {command}
          </div>
        ))}
      </div>
    </div>
  )
}

function extractCommandHints(content: string) {
  const matches = content.match(/`(?:npx|npm|pnpm|yarn|uvx|uv|python -m)[^`\n]*`/g) ?? []
  const normalized = matches.map((entry) => entry.replaceAll("`", "").trim())
  return Array.from(new Set(normalized)).slice(0, 6)
}

function formatInstalledAt(value: number) {
  return new Date(value).toLocaleString()
}
