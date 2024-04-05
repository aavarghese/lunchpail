type KubernetesResource<
  ApiVersion extends string = string,
  Kind extends string = string,
  Spec = unknown,
  Annotations = unknown,
  Labels = unknown,
  Top = unknown,
> = Top & {
  apiVersion: ApiVersion
  kind: Kind

  /** Resource metadata */
  metadata: Labels & {
    /** Resource name */
    name: string

    /** Resource namespace */
    namespace: string

    /** The cluster in which this resources resides */
    context: string

    /** Age of resource */
    creationTimestamp: string

    /** Resource annotations */
    annotations: Annotations & {
      /** Status of Resource (TODO) */
      "lunchpail.io/status": string

      /** Coded reason for failure (TODO) */
      "lunchpail.io/reason"?: string

      /** Error message (TODO) */
      "lunchpail.io/message"?: string
    }
  }

  /** Resource spec */
  spec: Readonly<Spec>
}

export type KubernetesSecret<Data> = KubernetesResource<"v1", "Secret", unknown, unknown, unknown, { data: Data }>
export type KubernetesS3Secret = KubernetesSecret<{ accessKeyID: string; secretAccessKey: string }>

export default KubernetesResource
