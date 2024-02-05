import CodeMirror from "@uiw/react-codemirror"
import { tags as t } from "@lezer/highlight"
import { EditorView } from "codemirror"
import { createTheme } from "@uiw/codemirror-themes"
import { useEffect, useMemo, useState } from "react"

// support for languages
import { json } from "@codemirror/lang-json"
import { python } from "@codemirror/lang-python"
import * as tclMode from "@codemirror/legacy-modes/mode/tcl"
import * as yamlMode from "@codemirror/legacy-modes/mode/yaml"
import * as shellMode from "@codemirror/legacy-modes/mode/shell"
import * as verilogMode from "@codemirror/legacy-modes/mode/verilog"
import { StreamLanguage, LanguageSupport } from "@codemirror/language"

import "./Code.css"

export type SupportedLanguage = "python" | "shell" | "json" | "yaml" | "verilog" | "tcl" | "makefile"

export type Props = {
  /** Content to display */
  children: string | Promise<string>

  /** How to interpret the syntax, for e.g. coloration */
  language: SupportedLanguage

  /** Should a left-side gutter with line numbers by displayed? [default=true] */
  showLineNumbers?: boolean

  /** Minimum height [default="250px"] */
  minHeight?: string

  /** Maximum height [default=no scroll] */
  maxHeight?: string

  /** Is the user allowed to change the content? [default=false] */
  readOnly?: boolean

  /** Callback when user changes content (ignored for `readOnly=false`) */
  onChange?: (val: string) => void
}

/** Which CodeMirror extensions should we use? */
function determineExtensionsToUse(props: Props) {
  const languageExtensions =
    props.language === "python"
      ? [python()]
      : props.language === "shell" || props.language === "makefile"
        ? [new LanguageSupport(StreamLanguage.define(shellMode.shell))]
        : props.language === "tcl"
          ? [new LanguageSupport(StreamLanguage.define(tclMode.tcl))]
          : props.language === "yaml"
            ? [new LanguageSupport(StreamLanguage.define(yamlMode.yaml))]
            : props.language === "verilog"
              ? [new LanguageSupport(StreamLanguage.define(verilogMode.verilog))]
              : [json()]

  return [...languageExtensions, ...(props.showLineNumbers !== false ? [EditorView.lineWrapping] : [])]
}

export default function Code(props: Props) {
  const [data, setData] = useState(typeof props.children === "string" ? props.children : "")

  useEffect(() => {
    if (typeof props.children !== "string") {
      props.children.then(setData)
    } else {
      setData(props.children)
    }
  }, [props.children, setData])

  // <CodeMirror/> doesn't call our `props.onChange` the first time,
  // though it will call that `onChange` on subsequent updates. Hence,
  // we need this "onMount" handler to push the initial value back
  useEffect(() => {
    if (props.onChange) {
      props.onChange(data)
    }
  }, [data])

  // which language extension do we want to use?
  const extensions = useMemo(() => determineExtensionsToUse(props), [props.language])

  return (
    <CodeMirror
      className="codeflare--code"
      data-show-line-numbers={String(props.showLineNumbers ?? true)}
      readOnly={props.readOnly ?? false}
      value={data}
      onChange={props.onChange}
      extensions={extensions}
      theme={patternflyTheme}
      minHeight={props.minHeight ?? "250px"}
      maxHeight={props.maxHeight}
    />
  )
}

const patternflyTheme = createTheme({
  theme: "dark",
  settings: {
    background: "var(--pf-v5-global--BackgroundColor--dark-100)",
    foreground: "var(--pf-v5-global--Color--light-100)",
    caret: "#c9d1d9",
    selection: "#2b9af350", // --pf-v5-global--palette--blue-50 with 50% opacity
    selectionMatch: "var(--pf-v5-global--palette--light-blue-300)",
    lineHighlight: "#F0F0F018", // var(--pf-v5-global--BackgroundColor--dark-400) with 18% opacity
  },
  styles: [
    { tag: [t.standard(t.tagName), t.tagName], color: "var(--pf-v5-global--palette--orange-200)" },
    { tag: [t.comment], color: "var(--pf-v5-global--palette--light-green-400)" },
    { tag: [t.bracket], color: "var(--pf-v5-global--palette--black-300)" },
    { tag: [t.className, t.propertyName], color: "var(--pf-v5-global--palette--light-blue-200)" },
    { tag: [t.variableName, t.attributeName, t.number, t.operator], color: "var(--pf-v5-global--palette--blue-100)" },
    { tag: [t.keyword, t.typeName, t.typeOperator, t.typeName], color: "var(--pf-v5-global--palette--blue-100)" },
    { tag: [t.meta], color: "var(--pf-v5-global--palette--light-blue-200)" },
    { tag: [t.string, t.regexp], color: "var(--pf-v5-global--palette--gold-100)" },
    { tag: [t.name, t.quote], color: "#7ee787" },
    { tag: [t.heading, t.strong], color: "#d2a8ff", fontWeight: "bold" },
    { tag: [t.emphasis], color: "#d2a8ff", fontStyle: "italic" },
    { tag: [t.deleted], color: "#ffdcd7", backgroundColor: "ffeef0" },
    { tag: [t.atom, t.bool, t.special(t.variableName)], color: "var(--pf-v5-global--palette--light-blue-300)" },
    { tag: t.link, textDecoration: "underline" },
    { tag: t.strikethrough, textDecoration: "line-through" },
    { tag: t.invalid, color: "#f97583" },
  ],
})
