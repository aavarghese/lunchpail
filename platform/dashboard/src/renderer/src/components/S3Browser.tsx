import { useEffect, useState, type KeyboardEvent, type MouseEvent } from "react"
import { Nav, MenuContent, MenuItem, MenuList, DrilldownMenu, Menu, Spinner, Text } from "@patternfly/react-core"

import type { BucketItem } from "@jay/common/api/s3"
import type DataSetEvent from "@jay/common/events/ModelDataEvent"
import type { KubernetesS3Secret } from "@jay/common/events/KubernetesResource"

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

/**
 * A React component that visualizes the forest given by `props.roots[]`.
 */
function NavBrowser(
  props: { roots: Tree[]; endpoint: string; bucket: string; accessKey: string; secretKey: string } & Pick<
    Required<typeof window.jay>,
    "s3"
  >,
) {
  const [menuDrilledIn, setMenuDrilledIn] = useState<string[]>([])
  const [drilldownPath, setDrilldownPath] = useState<string[]>([])
  const [menuHeights, setMenuHeights] = useState<Record<string, number>>({})
  const [activeMenu, setActiveMenu] = useState("nav-drilldown-rootMenu")

  const onDrillIn = (_event: KeyboardEvent | MouseEvent, fromItemId: string, toItemId: string, itemId: string) => {
    setMenuDrilledIn((prevMenuDrilledIn) => [...prevMenuDrilledIn, fromItemId])
    setDrilldownPath((prevDrilldownPath) => [...prevDrilldownPath, itemId])
    setActiveMenu(toItemId)
  }

  const onDrillOut = (_event: KeyboardEvent | MouseEvent, toItemId: string /*, _itemId: string*/) => {
    setMenuDrilledIn((prevMenuDrilledIn) => prevMenuDrilledIn.slice(0, prevMenuDrilledIn.length - 1))
    setDrilldownPath((prevDrilldownPath) => prevDrilldownPath.slice(0, prevDrilldownPath.length - 1))
    setActiveMenu(toItemId)
  }

  const onGetMenuHeight = (menuId: string, height: number) => {
    if (
      (menuHeights[menuId] !== height && menuId !== "nav-drilldown-rootMenu") ||
      (!menuHeights[menuId] && menuId === "nav-drilldown-rootMenu")
    ) {
      setMenuHeights((prevMenuHeights) => ({ ...prevMenuHeights, [menuId]: height }))
    }
  }

  function showContent(item: Tree) {
    const [loading, setLoading] = useState(true)
    const [content, setContent] = useState<string | { error: unknown } | null>(null)

    useEffect(() => {
      async function fetch() {
        try {
          const { accessKey, secretKey, endpoint, bucket } = props
          const content = await props.s3.getObject(endpoint, accessKey, secretKey, bucket, item.name)

          setContent(content)
        } catch (error) {
          setContent({ error })
        }

        setLoading(false)
      }

      fetch()
    }, [item.name, props.endpoint, props.bucket, props.accessKey, props.secretKey])

    const description =
      loading || !content ? (
        <Spinner />
      ) : isError(content) ? (
        "Error loading content: " + content.error
      ) : (
        viewContent(content, item.name)
      )
    return (
      <MenuItem
        key={item.name}
        itemId={`s3nav-content`}
        description={description}
        className="codeflare--no-hover"
      ></MenuItem>
    )
  }

  function toMenuItems(roots: Tree[], depth: number, parent?: Tree) {
    const baseId = `s3nav-drilldown-${depth}-`

    return [
      ...(!parent
        ? []
        : [
            <MenuItem key="up" itemId={`${baseId}-up`} direction="up">
              {parent.name}
            </MenuItem>,
          ]),
      ...(roots.length === 0 && parent && hasViewableContent(parent.name) ? [showContent(parent)] : []),
      ...roots.map((item, idx) => (
        <MenuItem
          key={item.name}
          itemId={baseId + `item-${idx}`}
          direction={!isLeafNode(item) || hasViewableContent(item.name) ? "down" : undefined}
          description={!isLeafNode(item) ? "Folder" : filetypeFromName(item.name)}
          drilldownMenu={
            <DrilldownMenu id={baseId + `drilldown-${idx}`}>
              {toMenuItems(item.children, depth + 1, item)}
            </DrilldownMenu>
          }
        >
          {item.name}
        </MenuItem>
      )),
    ]
  }

  return (
    <Nav aria-label="s3 file browser" className="codeflare--s3-browser">
      <Menu
        id="s3nav-drilldown-rootMenu"
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
 * `Tree` model.
 */
export default function S3Browser(
  props: DataSetEvent["spec"]["local"] & Pick<Required<typeof window.jay>, "get" | "s3">,
) {
  const [loading, setLoading] = useState(true)
  const [secret, setSecret] = useState<null | { accessKey: string; secretKey: string }>(null)
  const [content, setContent] = useState<BucketItem[] | { error: unknown } | null>(null)

  useEffect(() => {
    async function fetch() {
      try {
        const secret = await props.get<KubernetesS3Secret>({
          kind: "secret",
          name: props["secret-name"],
          namespace: props["secret-namespace"],
        })

        const accessKey = atob(secret.data.accessKeyID)
        const secretKey = atob(secret.data.secretAccessKey)
        const items = await props.s3.listObjects(props.endpoint, accessKey, secretKey, props.bucket)

        setSecret({ accessKey, secretKey })
        setContent(items)
      } catch (error) {
        setContent({ error })
      }

      setLoading(false)
    }

    fetch()
  }, [props["secret-name"], props["secret-namespace"]])

  if (loading || content === null || secret === null) {
    return <Spinner />
  } else if (isError(content)) {
    console.error("Error loading secrets", content)
    return "Error loading secrets: " + content.error
  } else {
    return (
      <NavBrowser roots={toTree(content)} {...secret} endpoint={props.endpoint} bucket={props.bucket} s3={props.s3} />
    )
  }
}

/**
 * Take a list of S3 objects and fold them into a `Tree` model based
 * on the `/` path separators in the `name` field of the `items`.
 */
function toTree(items: BucketItem[]): Tree[] {
  const slashes = /\//
  return items.reduce(
    (r, s) => {
      if (s.name) {
        s.name.split(slashes).reduce((q, _, i, a) => {
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
  md: "Markdown",
  json: "JSON",
  txt: "Text",
}

function filetypeFromName(name: string) {
  const extIdx = name.lastIndexOf(".")
  if (extIdx >= 0) {
    const ext = name.slice(extIdx + 1)
    return filetypeLookup[ext] || undefined
  } else {
    return undefined
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
  if (/text/i.test(ext)) {
    return <Text component="pre">{content}</Text>
  } else if (/json/i.test(ext)) {
    return <Text component="pre">{JSON.stringify(JSON.parse(content), undefined, 2)}</Text>
  } else {
    return <Text component="pre">{content}</Text>
  }
}
