import type WatchedKind from "@jay/common/Kind"
import type ContentProvider from "./ContentProvider"

import datasets from "./datasets"
import taskqueues from "./taskqueues"
import workerpools from "./workerpools"
import controlplane from "./controlplane"
import applications from "./applications"
import platformreposecrets from "./platformreposecrets"

export type { ContentProvider }

/**
 * These are the resource Kinds for which we have UI componetry.
 */
const uiProviders = {
  controlplane,
  platformreposecrets,
  applications,
  taskqueues,
  datasets,
  workerpools,
}

/** Do we have a Detail view? */
export type DetailableKind = keyof typeof uiProviders

/** Do we have a Detail view? */
export function isDetailableKind(kind: WatchedKind | DetailableKind): kind is DetailableKind {
  return kind in uiProviders
}

/** Special cases: any Kinds we have Detail views for, but we want to exclude from the Nav UI? */
export type NavigableKind = Exclude<DetailableKind, "taskqueues">

/** Special cases: any Kinds we have Detail views for, but we want to exclude from the Nav UI? */
export function isNavigableKind(kind: WatchedKind | NavigableKind): kind is NavigableKind {
  return uiProviders[kind] && !!uiProviders[kind].isInSidebar
}

export default uiProviders