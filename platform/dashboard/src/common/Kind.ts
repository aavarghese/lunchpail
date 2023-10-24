/** Resource types that we watch */
const watchedKinds = ["datasets", "queues", "workerpools", "applications", "platformreposecrets"] as const

/** Valid resource types */
type WatchedKind = (typeof watchedKinds)[number]

export default WatchedKind
