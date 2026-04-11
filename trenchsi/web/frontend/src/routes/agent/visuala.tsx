import { createFileRoute } from "@tanstack/react-router"

import { VisualaPage } from "@/components/visuala/visuala-page"

export const Route = createFileRoute("/agent/visuala")({
  component: AgentVisualaRoute,
})

function AgentVisualaRoute() {
  return <VisualaPage />
}
