// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

import XLSX from "xlsx-style"

const ID = "ID"
const Description = "Description"
const Default = "Default message"

/**
 * parse an xlsx document into a locales object.
 *
 * @param {Buffer} buf - The buffer to parse.
 * @returns {object} - The locales
 */
const unmarshal = function (buf) {
  const xlsx = XLSX.read(buf)

  const sheet = xlsx.Sheets.translations || xlsx.Sheets.Sheet1

  const messages =
    Object.keys(sheet)
      .filter(key => key[0] !== "!")
      .map(function (addr) {
        const cell = sheet[addr]
        const colHeaderAddr = addr.replace(/[A-Z]+/, "A")
        const rowHeaderAddr = addr.replace(/[0-9]+/, "1")

        const colHeader = sheet[colHeaderAddr]
        const rowHeader = sheet[rowHeaderAddr]

        if (!colHeader || !rowHeaderAddr) {
          return null
        }

        const id = colHeader.v
        const header = rowHeader.v.toLowerCase().trim()
        switch (header) {
        case ID.toLowerCase():
          return null
        case Description.toLowerCase():
          return { id, description: cell.v }
        case Default.toLowerCase():
          return { id, defaultMessage: cell.v }
        default:
          return { id, [header]: cell.v }
        }
      })
      .filter(message => Boolean(message))
      .reduce(function (acc, message) {
        const merged = acc[message.id] || { id: message.id, locales: {}}

        Object.keys(message).forEach(function (key) {
          if (key === "id" || key === "defaultMessage" || key === "description") {
            merged[key] = message[key]
            return
          }
          merged.locales[key] = message[key]
        })

        acc[message.id] = merged
        return acc
      }, {})

  return messages
}

/**
 * stringify messages into an xlsx document.
 *
 * @param {object} messages - The messages to marshal.
 * @returns {string} - The marshalled xlsx document.
 */
const marshal = function (messages) {
  const Workbook = function () {
    if (!(this instanceof Workbook)) {
      return new Workbook()
    }

    this.SheetNames = []
    this.Sheets = {}
  }

  const wb = new Workbook()
  const ws = {}

  const locales = messages[0] ? Object.keys(messages[0].locales).sort() : []
  const titles = [ ID, Description, Default, ...locales ]

  let maxCol = 0
  let maxRow = 0

  const titleStyle = {
    font: {
      bold: true,
    },
    border: {
      bottom: {
        style: "medium",
        color: { auto: "1" },
      },
    },
    alignment: {
      vertical: "center",
      horizontal: "left",
    },
  }

  const emptyStyle = { fill: { fgColor: { rgb: "fec7ce" }}}

  titles.forEach(function (name, i) {
    maxCol = Math.max(maxCol, i)
    const addr = XLSX.utils.encode_cell({
      r: 0,
      c: i,
    })
    ws[addr] = {
      t: "s",
      v: name,
      s: titleStyle,
    }
  })

  ws["!cols"] = [
    {
      width: "25",
      style: "1",
      customWidth: "1",
      wpx: 100,
      wch: 24.17,
      MDW: 6,
    },
    {
      width: "25",
      style: "1",
      customWidth: "1",
      wpx: 200,
      wch: 24.17,
      MDW: 6,
    },
    {
      width: "25",
      style: "1",
      customWidth: "1",
      wpx: 200,
      wch: 24.17,
      MDW: 6,
    },
  ]

  Object.keys(messages).forEach(function (id, j) {
    const message = messages[id]
    message.id = id
    const r = j + 1
    maxRow = Math.max(maxRow, r)

    // set id
    ws[XLSX.utils.encode_cell({ c: 0, r })] = {
      t: "s",
      v: message.id,
    }

    // set description
    ws[XLSX.utils.encode_cell({ c: 1, r })] = {
      t: "s",
      v: message.description,
    }

    // set default message
    ws[XLSX.utils.encode_cell({ c: 2, r })] = {
      t: "s",
      v: message.defaultMessage,
    }

    Object.keys(message.locales).sort().forEach(function (locale, i) {
      const k = 3 + i
      maxCol = Math.max(maxCol, k)
      // set locale
      const v = message.locales[locale]
      ws[XLSX.utils.encode_cell({ c: k, r })] = {
        t: "s",
        v,
        s: v === "" ? emptyStyle : {},
      }

      ws[XLSX.utils.encode_cell({ c: k, r: 0 })] = {
        t: "s",
        v: locale,
        s: titleStyle,
      }

      ws["!cols"][k] = {
        width: "85",
        bestFit: "1",
        customWidth: "1",
        wpx: 300,
        wch: 85,
        MDW: 6,
      }
    })
  })

  ws["!ref"] = XLSX.utils.encode_range({
    s: { c: 0, r: 0 },
    e: { c: maxCol, r: maxRow },
  })

  wb.Sheets.translations = ws
  wb.SheetNames = [ "translations" ]

  const r = XLSX.write(wb, {
    bookType: "xlsx",
    bookSST: true,
    type: "buffer",
  })

  return r
}

export default {
  unmarshal,
  marshal,
}
