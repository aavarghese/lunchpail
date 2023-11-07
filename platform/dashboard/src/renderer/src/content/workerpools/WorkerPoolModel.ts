/** Map from TaskQueue label to number of tasks to be done for that TaskQueue */
export type TaskQueueTask = Record<string, number>

/** One TaskQueueTask for each Worker in a WorkerPool */
type TasksAcrossWorkers = TaskQueueTask[]

export interface WorkerPoolModel {
  inbox: TasksAcrossWorkers
  outbox: TasksAcrossWorkers
  processing: TasksAcrossWorkers
  label: string

  /** Namespace in which this model is stored */
  namespace: string
}

export type WorkerPoolModelWithHistory = WorkerPoolModel & {
  numEvents: number
  events: { outbox: number; timestamp: number }[]
}