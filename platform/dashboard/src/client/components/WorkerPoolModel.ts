/** Map from DataSet label to number of tasks to be done for that DataSet */
export type DataSetTask = Record<string, number>

/** One DataSetTask for each Worker in a WorkerPool */
type TasksAcrossWorkers = DataSetTask[]

export interface WorkerPoolModel {
  inbox: TasksAcrossWorkers
  outbox: TasksAcrossWorkers
  processing: TasksAcrossWorkers
  label: string
}

export interface WorkerPoolModelWithHistory extends WorkerPoolModel {
  outboxHistory: number[]
  timestamps: number[]
}

export default interface QueueEvent {
  /** millis since epoch */
  timestamp: number

  inbox: number
  outbox: number
  processing: number
  dataset: string
  workerpool: string
  workerIndex: number
}
