import { createFileRoute } from "@tanstack/react-router"

import { LearnedSkillsPage } from "@/components/skills/learned-skills-page"

export const Route = createFileRoute("/agent/learned")({
  component: AgentLearnedRoute,
})

function AgentLearnedRoute() {
  return <LearnedSkillsPage />
}
