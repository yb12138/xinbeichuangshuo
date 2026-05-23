/// <reference types="node" />

import { describe, expect, it } from 'vitest'
import ts from 'typescript'
import { readdirSync, readFileSync } from 'node:fs'
import path from 'node:path'

const scenarioRoot = path.resolve(process.cwd(), 'e2e/scenarios')
const controlOptionIDs = new Set(['-1', 'cancel', 'decline', 'refuse', 'skip', 'pass', 'back', 'done', 'finish'])

type OptionNode = {
  node: ts.ObjectLiteralExpression
  context: string
}

function propName(name: ts.PropertyName): string {
  if (ts.isIdentifier(name) || ts.isStringLiteral(name) || ts.isNumericLiteral(name)) return name.text
  return ''
}

function getProperty(node: ts.ObjectLiteralExpression, name: string): ts.ObjectLiteralElementLike | undefined {
  return node.properties.find((prop) => {
    if (ts.isPropertyAssignment(prop)) return propName(prop.name) === name
    if (ts.isShorthandPropertyAssignment(prop)) return prop.name.text === name
    return false
  })
}

function getPropertyValue(node: ts.ObjectLiteralExpression, name: string): ts.Expression | undefined {
  const prop = getProperty(node, name)
  if (!prop) return undefined
  if (ts.isPropertyAssignment(prop)) return prop.initializer
  if (ts.isShorthandPropertyAssignment(prop)) return prop.name
  return undefined
}

function hasProperty(node: ts.ObjectLiteralExpression, name: string): boolean {
  return getProperty(node, name) != null
}

function unwrapExpression(expression: ts.Expression): ts.Expression {
  let current = expression
  while (ts.isAsExpression(current) || ts.isSatisfiesExpression(current) || ts.isParenthesizedExpression(current)) {
    current = current.expression
  }
  return current
}

function stringLiteralValue(expression: ts.Expression | undefined): string | undefined {
  if (!expression) return undefined
  const unwrapped = unwrapExpression(expression)
  return ts.isStringLiteralLike(unwrapped) ? unwrapped.text : undefined
}

function containsNode(node: ts.Node, predicate: (candidate: ts.Node) => boolean): boolean {
  if (predicate(node)) return true
  let found = false
  node.forEachChild((child) => {
    if (!found && containsNode(child, predicate)) found = true
  })
  return found
}

function isDisallowedTargetIDFallback(expression: ts.Expression): boolean {
  return containsNode(unwrapExpression(expression), (node) => {
    if (ts.isBinaryExpression(node) && (node.operatorToken.kind === ts.SyntaxKind.QuestionQuestionToken || node.operatorToken.kind === ts.SyntaxKind.BarBarToken)) {
      return true
    }
    if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'String') {
      return true
    }
    if (ts.isIdentifier(node) && (node.text === 'idx' || node.text === 'index')) {
      return true
    }
    return false
  })
}

function isTargetPickerPrompt(node: ts.ObjectLiteralExpression): boolean {
  const presentation = getPropertyValue(node, 'presentation')
  if (!presentation) return false
  const presentationObject = unwrapExpression(presentation)
  if (!ts.isObjectLiteralExpression(presentationObject)) return false
  return stringLiteralValue(getPropertyValue(presentationObject, 'kind')) === 'target_picker'
}

function hasTargetFilter(node: ts.ObjectLiteralExpression): boolean {
  const presentation = getPropertyValue(node, 'presentation')
  if (!presentation) return false
  const presentationObject = unwrapExpression(presentation)
  return ts.isObjectLiteralExpression(presentationObject) && hasProperty(presentationObject, 'target_filter')
}

function nearestFunctionScope(node: ts.Node): ts.Node {
  let current: ts.Node | undefined = node.parent
  while (current) {
    if (ts.isFunctionLike(current) || ts.isSourceFile(current)) return current
    current = current.parent
  }
  return node.getSourceFile()
}

function objectFromMapInitializer(expression: ts.Expression): ts.ObjectLiteralExpression | undefined {
  const unwrapped = unwrapExpression(expression)
  if (!ts.isCallExpression(unwrapped) || !ts.isPropertyAccessExpression(unwrapped.expression)) return undefined
  if (unwrapped.expression.name.text !== 'map') return undefined
  const callback = unwrapped.arguments[0]
  if (!callback || (!ts.isArrowFunction(callback) && !ts.isFunctionExpression(callback))) return undefined
  const body = callback.body
  if (ts.isObjectLiteralExpression(body)) return body
  if (ts.isParenthesizedExpression(body) && ts.isObjectLiteralExpression(body.expression)) return body.expression
  if (ts.isBlock(body)) {
    for (const statement of body.statements) {
      if (!ts.isReturnStatement(statement) || !statement.expression) continue
      const returned = unwrapExpression(statement.expression)
      if (ts.isObjectLiteralExpression(returned)) return returned
    }
  }
  return undefined
}

function collectIdentifierOptions(scope: ts.Node, name: string): OptionNode[] {
  const result: OptionNode[] = []

  function visit(node: ts.Node) {
    if (ts.isFunctionLike(node) && node !== scope) return

    if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === name && node.initializer) {
      const initializer = unwrapExpression(node.initializer)
      if (ts.isArrayLiteralExpression(initializer)) {
        for (const element of initializer.elements) {
          const option = unwrapExpression(element)
          if (ts.isObjectLiteralExpression(option)) result.push({ node: option, context: `${name} initializer` })
        }
      } else {
        const mapped = objectFromMapInitializer(initializer)
        if (mapped) result.push({ node: mapped, context: `${name} map initializer` })
      }
    }

    if (
      ts.isCallExpression(node) &&
      ts.isPropertyAccessExpression(node.expression) &&
      node.expression.name.text === 'push' &&
      ts.isIdentifier(node.expression.expression) &&
      node.expression.expression.text === name
    ) {
      for (const argument of node.arguments) {
        const option = unwrapExpression(argument)
        if (ts.isObjectLiteralExpression(option)) result.push({ node: option, context: `${name}.push` })
      }
    }

    ts.forEachChild(node, visit)
  }

  visit(scope)
  return result
}

function optionNodesForPrompt(prompt: ts.ObjectLiteralExpression): OptionNode[] | null {
  const options = getPropertyValue(prompt, 'options')
  if (!options) return null
  const initializer = unwrapExpression(options)
  if (ts.isArrayLiteralExpression(initializer)) {
    return initializer.elements.flatMap((element) => {
      const option = unwrapExpression(element)
      return ts.isObjectLiteralExpression(option) ? [{ node: option, context: 'inline options' }] : []
    })
  }
  if (ts.isIdentifier(initializer)) {
    return collectIdentifierOptions(nearestFunctionScope(prompt), initializer.text)
  }
  return null
}

function sourceLocation(sourceFile: ts.SourceFile, node: ts.Node): string {
  const pos = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile))
  return `${path.relative(scenarioRoot, sourceFile.fileName)}:${pos.line + 1}:${pos.character + 1}`
}

describe('target picker scenario protocol', () => {
  it('requires explicit target_id and target_filter in e2e scenario prompts', () => {
    const issues: string[] = []
    let promptCount = 0
    let optionCount = 0

    for (const fileName of readdirSync(scenarioRoot).filter((name) => name.endsWith('.ts'))) {
      const filePath = path.join(scenarioRoot, fileName)
      const sourceFile = ts.createSourceFile(filePath, readFileSync(filePath, 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)

      function visit(node: ts.Node) {
        if (ts.isObjectLiteralExpression(node) && isTargetPickerPrompt(node)) {
          promptCount += 1
          if (!hasTargetFilter(node)) issues.push(`${sourceLocation(sourceFile, node)} missing presentation.target_filter`)

          const options = optionNodesForPrompt(node)
          if (!options || options.length === 0) {
            issues.push(`${sourceLocation(sourceFile, node)} could not resolve target_picker options`)
          } else {
            for (const option of options) {
              optionCount += 1
              const targetID = getPropertyValue(option.node, 'target_id')
              const id = stringLiteralValue(getPropertyValue(option.node, 'id'))
              if (!targetID && (!id || !controlOptionIDs.has(id.trim().toLowerCase()))) {
                issues.push(`${sourceLocation(sourceFile, option.node)} ${option.context} option is missing target_id`)
              } else if (targetID && isDisallowedTargetIDFallback(targetID)) {
                issues.push(`${sourceLocation(sourceFile, targetID)} ${option.context} target_id uses a fallback expression`)
              }
            }
          }
        }
        ts.forEachChild(node, visit)
      }

      visit(sourceFile)
    }

    expect(promptCount).toBeGreaterThan(0)
    expect(optionCount).toBeGreaterThan(0)
    expect(issues).toEqual([])
  })
})
