import detail from "./detail"
import { name, singular } from "./name"

import type ContentProvider from "../ContentProvider"

const taskqueues: ContentProvider<"taskqueues"> = {
  kind: "taskqueues" as const,
  name,
  singular,
  detail,
  description: "not needed",
}

export default taskqueues
