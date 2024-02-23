import untruncateJson from "untruncate-json"
import { useCallback, useEffect, useState, type KeyboardEvent, type MouseEvent } from "react"
import { Badge, Nav, MenuContent, MenuItem, MenuList, DrilldownMenu, Menu, Spinner, Text } from "@patternfly/react-core"

import Code from "./Code"
import Json from "./Json"

import type { BucketItem } from "@jaas/common/api/s3"
import type DataSetEvent from "@jaas/common/events/DataSetEvent"
import type TaskQueueEvent from "@jaas/common/events/TaskQueueEvent"
import type { KubernetesS3Secret } from "@jaas/common/events/KubernetesResource"

import "./S3Browser.scss"

type InteriorNode = {
  name: string
  children: Tree[]
}

type LeafNode = InteriorNode & Pick<BucketItem, "lastModified">

type Tree = LeafNode | InteriorNode

function isLeafNode(item: Tree): item is LeafNode {
  const node = item as InteriorNode
  return typeof node.name === "string" && (!Array.isArray(node.children) || node.children.length === 0)
}

type PathPrefix = {
  /**
   * S3 folder prefix to browse. If not specified, the root of the bucket will be displayed.
   */
  prefix: string
}

type S3Props = { endpoint: string; bucket: string; accessKey: string; secretKey: string } & Pick<
  Required<typeof window.jaas>,
  "s3"
>
type NavBrowserProps = S3Props & { roots: Tree[] } & Partial<PathPrefix>

/**
 * A React component that visualizes the forest given by
 * `props.roots[]` by using a PatternFly `<Nav/>` with its drilldown
 * feature.
 */
function NavBrowser(props: NavBrowserProps) {
  const rootMenuId = "s3nav-drilldown-rootMenu"
  const [menuDrilledIn, setMenuDrilledIn] = useState<string[]>([])
  const [drilldownPath, setDrilldownPath] = useState<string[]>([])
  const [menuHeights, setMenuHeights] = useState<Record<string, number>>({})
  const [activeMenu, setActiveMenu] = useState(rootMenuId)

  const onDrillIn = useCallback(
    (_event: KeyboardEvent | MouseEvent, fromItemId: string, toItemId: string, itemId: string) => {
      setMenuDrilledIn((prevMenuDrilledIn) => [...prevMenuDrilledIn, fromItemId])
      setDrilldownPath((prevDrilldownPath) => [...prevDrilldownPath, itemId])
      setActiveMenu(toItemId)
    },
    [],
  )

  const onDrillOut = useCallback((_event: KeyboardEvent | MouseEvent, toItemId: string /*, _itemId: string*/) => {
    setMenuDrilledIn((prevMenuDrilledIn) => prevMenuDrilledIn.slice(0, prevMenuDrilledIn.length - 1))
    setDrilldownPath((prevDrilldownPath) => prevDrilldownPath.slice(0, prevDrilldownPath.length - 1))
    setActiveMenu(toItemId)
  }, [])

  const onGetMenuHeight = useCallback((menuId: string, height: number) => {
    //if ((menuHeights[menuId] !== height && menuId !== rootMenuId) || (!menuHeights[menuId] && menuId === rootMenuId)) {
    setMenuHeights((prevMenuHeights) => {
      if (
        (prevMenuHeights[menuId] !== height && menuId !== rootMenuId) ||
        (!prevMenuHeights[menuId] && menuId === rootMenuId)
      ) {
        if (height !== 1 && prevMenuHeights[menuId] !== height) {
          // without this check, the patternfly component enters an infinite loop of e.g. 145->1, 1->145, ...
          return { ...prevMenuHeights, [menuId]: height }
        }
      }

      return prevMenuHeights
    })
  }, [])

  function toMenuItems(roots: Tree[], depth: number, parent?: Tree, parentMenuId?: string) {
    const baseId = `s3nav-drilldown-${depth}-`

    return [
      ...(!parent
        ? []
        : [
            <MenuItem key="up" itemId={`${baseId}-up`} direction="up">
              {parent.name}
            </MenuItem>,
          ]),
      ...(roots.length === 0 && parent && parentMenuId && hasViewableContent(parent.name) && activeMenu === parentMenuId
        ? [<ShowContent key={parent.name} object={parent.name} {...props} />]
        : []),
      ...roots.map((item, idx) => {
        const drilldownMenuId = baseId + `drilldown-${idx}`

        const childFilter = new RegExp("^" + item.name + "/$")
        const children = item.children.filter((_) => !childFilter.test(_.name))

        return (
          <MenuItem
            key={item.name}
            itemId={baseId + `item-${idx}`}
            direction={!isLeafNode(item) || hasViewableContent(item.name) ? "down" : undefined}
            description={!isLeafNode(item) ? "Folder" : filetypeFromName(item.name)}
            drilldownMenu={
              <DrilldownMenu id={drilldownMenuId}>
                {toMenuItems(children, depth + 1, item, drilldownMenuId)}
              </DrilldownMenu>
            }
          >
            {parent ? item.name.replace(parent.name + "/", "") : item.name}{" "}
            {!isLeafNode(item) && <Badge>{children.length}</Badge>}
          </MenuItem>
        )
      }),
    ]
  }

  return (
    <Nav aria-label="s3 file browser" className="codeflare--s3-browser">
      <Menu
        id={rootMenuId}
        containsDrilldown
        drilldownItemPath={drilldownPath}
        drilledInMenus={menuDrilledIn}
        activeMenu={activeMenu}
        onDrillIn={onDrillIn}
        onDrillOut={onDrillOut}
        onGetMenuHeight={onGetMenuHeight}
      >
        <MenuContent menuHeight={menuHeights[activeMenu] ? `${menuHeights[activeMenu]}px` : undefined}>
          <MenuList>{toMenuItems(props.roots, 0)}</MenuList>
        </MenuContent>
      </Menu>
    </Nav>
  )
}

/**
 * A React component that presents a `<NavBrowser/>` after loading the
 * `Tree` model. This component will manage fetching the S3
 * credentials associated with the given `DataSet`, and then pass them
 * to `<S3BrowserWithCreds/>`.
 */
function S3Browser(
  props: DataSetEvent["spec"]["local"] & Pick<Required<typeof window.jaas>, "get" | "s3"> & Partial<PathPrefix>,
) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<unknown | null>(null)
  const [secret, setSecret] = useState<null | { accessKey: string; secretKey: string }>(null)

  useEffect(() => {
    async function fetchCredentials() {
      try {
        const secret = await props.get<KubernetesS3Secret>({
          kind: "secret",
          name: props["secret-name"],
          namespace: props["secret-namespace"],
        })

        const accessKey = atob(secret.data.accessKeyID)
        const secretKey = atob(secret.data.secretAccessKey)
        setSecret({ accessKey, secretKey })
      } catch (error) {
        console.error("Error fetching S3 credentials", props, error)
        setError(error)
      }

      setLoading(false)
    }

    fetchCredentials()
  }, [props["secret-name"], props["secret-namespace"]])

  if (loading || secret === null) {
    return <Spinner />
  } else if (error) {
    return "Error loading secrets: " + error
  } else {
    return (
      <S3BrowserWithCreds
        {...secret}
        endpoint={props.endpoint}
        bucket={props.bucket}
        s3={props.s3}
        prefix={props.prefix}
      />
    )
  }
}

export function S3BrowserWithCreds(props: Omit<NavBrowserProps, "roots"> & Partial<PathPrefix>) {
  const [loading, setLoading] = useState(true)
  const [content, setContent] = useState<BucketItem[] | { error: unknown } | null>(null)

  useEffect(() => {
    async function fetchContent() {
      try {
        const { accessKey, secretKey, endpoint, bucket, prefix } = props
        const items = await props.s3.listObjects(endpoint, accessKey, secretKey, bucket, prefix)
        setContent(items)
      } catch (error) {
        console.error("Error listing S3 objects", props, error)
        setContent({ error })
      } finally {
        setLoading(false)
      }
    }

    // TODO: polling... can we do better? add a refresh button somehow?
    fetchContent()
    const interval = setInterval(fetchContent, 5000)

    // return the cleanup function to react; it will call this on
    // component unmount
    return () => clearInterval(interval)
  }, [props.endpoint, props.accessKey, props.secretKey, props.bucket, setContent, setLoading])

  if (loading || content === null) {
    return <Spinner />
  } else if (isError(content)) {
    console.error("Error loading secrets", content)
    return "Error loading secrets: " + content.error
  } else if (content.length === 0) {
    console.log("No S3 objects found", props)
    return <span style={hasPadding}>No objects found for bucket {props.bucket}</span>
  } else {
    return (
      <NavBrowser
        roots={toTree(content, props.prefix)}
        endpoint={props.endpoint}
        accessKey={props.accessKey}
        secretKey={props.secretKey}
        bucket={props.bucket}
        s3={props.s3}
        prefix={props.prefix}
      />
    )
  }
}

/** Simulate hasPadding for DrawerTab */
const hasPadding = {
  paddingBlockStart: "var(--pf-v5-c-drawer--child--PaddingTop)",
  paddingBlockEnd: "var(--pf-v5-c-drawer--child--PaddingBottom)",
  paddingInlineStart: "var(--pf-v5-c-drawer--child--PaddingLeft)",
  paddingInlineEnd: "var(--pf-v5-c-drawer--child--PaddingRight)",
}

/**
 * Take a list of S3 objects and fold them into a `Tree` model based
 * on the `/` path separators in the `name` field of the `items`.
 */
function toTree(items: BucketItem[], prefix?: string): Tree[] {
  const slashes = /\//
  const prefixPattern = prefix ? new RegExp("^" + prefix + (prefix.endsWith("/") ? "" : "/")) : undefined

  return items
    .slice(0, 200)
    .map((_) => (!prefixPattern || !_.name ? _.name : _.name.replace(prefixPattern, "")))
    .reduce(
      (r, name) => {
        if (name) {
          name.split(slashes).reduce((q, _, i, a) => {
            const name = a.slice(0, i + 1).join("/")
            let existingChild = (q.children = q.children || []).find((o) => o.name === name)
            if (!existingChild) q.children.push((existingChild = { name, children: [] }))
            return existingChild
          }, r)
        }
        return r
      },
      { children: [] as Tree[] },
    ).children
}

function isError(response: null | unknown | { error: unknown }): response is { error: unknown } {
  return response !== null && typeof response === "object" && "error" in response
}

const filetypeLookup = {
  py: "Python",
  md: "Markdown",
  json: "JSON",
  txt: "Text",
  mk: "Makefile",
  Makefile: "Makefile",
  sdc: "Synopsys Design Constraint",
  v: "Verilog",
  gitignore: "Text",
  tcl: "TCL",
  cfg: "Text",
}

function filetypeFromName(name: string) {
  const extIdx = name.lastIndexOf(".")
  if (extIdx >= 0) {
    const ext = name.slice(extIdx + 1)
    return filetypeLookup[ext] || undefined
  } else {
    // maybe we have an entry for the whole name?
    return filetypeLookup[name.slice(name.lastIndexOf("/") + 1)]
  }
}

/**
 * This is imperfect: i.e. if `filetypeLookup` has a mapping for the
 * file extension of the file with `name`, then it is viewable by
 * us, as in the `viewContent()` function knows what to do.
 */
function hasViewableContent(name: string) {
  return !!filetypeFromName(name)
}

/**
 * @return a React component to visualize the given `content` for the
 * S3 `objectName`
 */
function viewContent(content: string, objectName: string) {
  const ext = filetypeFromName(objectName)
  if (/^(makefile|tcl|markdown|verilog|synopsys|py)/i.test(ext)) {
    return (
      <Code readOnly language={ext.toLowerCase()}>
        {content}
      </Code>
    )
  } else if (/json/i.test(ext)) {
    // the Menu bits give us the padding, so we don't need extra
    // padding from the Json viewer
    try {
      return <Json readOnly>{JSON.stringify(JSON.parse(content), undefined, 2)}</Json>
    } catch (err) {
      console.error("Error parsing JSON", err)

      // try to rectify it, perhaps this is truncated JSON?
      try {
        const rectified = JSON.parse(untruncateJson(content))
        rectified["warning"] = "This object has been truncated"
        return <Json readOnly>{JSON.stringify(rectified, undefined, 2)}</Json>
      } catch (err) {
        console.error("Error trying to rectify partial JSON", err)
      }
      // intentional fall-through
    }
  }

  return <Text component="pre">{content}</Text>
}

/**
 * Hijack a MenuItem to display content
 */
function ShowContent(props: S3Props & Partial<PathPrefix> & { object: string }) {
  const { s3, endpoint, accessKey, secretKey, bucket } = props

  const [loading, setLoading] = useState(true)
  const [content, setContent] = useState<string | { error: unknown } | null>(null)

  // on mount, we fetch the content
  useEffect(() => {
    async function fetch() {
      try {
        console.log("Fetching S3 content", props)
        setLoading(true)
        const content = await s3.getObject(
          endpoint,
          accessKey,
          secretKey,
          bucket,
          join(props.prefix ?? "", props.object),
        )
        console.log("Successfully fetched S3 content", props)
        setContent(content)
      } catch (error) {
        console.error("Error fetching S3 content", props, error)
        setContent({ error })
      } finally {
        setLoading(false)
      }
    }

    fetch()
  }, [props.object, endpoint, bucket, accessKey, secretKey])

  // we use the MenuItem `description` to house the view of the content
  const description =
    loading || !content ? (
      <Spinner />
    ) : isError(content) ? (
      "Error loading content: " + content.error
    ) : (
      viewContent(content, props.object)
    )

  return (
    <MenuItem
      key={props.object}
      itemId={`s3nav-content-${props.object}`}
      description={description}
      className="codeflare--no-hover codeflare--menu-item-as-content"
    ></MenuItem>
  )
}

import DrawerTab from "@jaas/components/Drawer/Tab"

/** A Drawer tab that shows <S3Browser /> */
export function BrowserTabs(
  props: (DataSetEvent | TaskQueueEvent)["spec"]["local"] & Partial<PathPrefix> & { title?: string },
) {
  if (window.jaas.get && window.jaas.s3) {
    return DrawerTab({
      hasNoPadding: true,
      title: props.title ?? "Browser",
      body: <S3Browser {...props} get={window.jaas.get} s3={window.jaas.s3} />,
    })
  } else {
    return undefined
  }
}

/** path.join */
function join(a: string, b: string) {
  return [a.replace(/\/$/, ""), b].join("/")
}
