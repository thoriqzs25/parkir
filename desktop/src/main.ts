import { app, BrowserWindow, ipcMain } from "electron";
import * as path from "path";

let mainWindow: BrowserWindow | null = null;
let cachedToken: string | null = null;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 1024,
    minHeight: 768,
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  mainWindow.loadFile(path.join(__dirname, "renderer", "index.html"));

  mainWindow.on("closed", () => {
    mainWindow = null;
  });
}

app.whenReady().then(createWindow);

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

app.on("activate", () => {
  if (mainWindow === null) {
    createWindow();
  }
});

// Token storage: in-memory in the main process for v1.
ipcMain.handle("get-token", () => {
  return cachedToken;
});

ipcMain.handle("set-token", (_event, token: string) => {
  cachedToken = token;
  return true;
});

ipcMain.handle("clear-token", () => {
  cachedToken = null;
  return true;
});

// Print helper: render arbitrary HTML in a hidden window and print it.
ipcMain.handle("print-html", async (_event, html: string) => {
  let printWindow: BrowserWindow | null = new BrowserWindow({
    show: false,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
    },
  });

  const completeHtml = `
    <!DOCTYPE html>
    <html>
      <head>
        <meta charset="UTF-8">
        <title>Print</title>
        <style>
          body { font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; font-size: 14px; }
          pre, .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
        </style>
      </head>
      <body>${html}</body>
    </html>
  `;

  const dataUrl = `data:text/html;charset=utf-8,${encodeURIComponent(completeHtml)}`;
  await printWindow.loadURL(dataUrl);

  return new Promise<void>((resolve, reject) => {
    if (!printWindow) {
      reject(new Error("print window was destroyed"));
      return;
    }

    const cleanup = () => {
      if (printWindow) {
        printWindow.destroy();
        printWindow = null;
      }
    };

    printWindow.webContents.on("did-finish-load", () => {
      if (!printWindow) return;
      printWindow.webContents.print(
        { silent: false, printBackground: false, copies: 1 },
        (success: boolean, failureReason: string) => {
          cleanup();
          if (success) {
            resolve();
          } else {
            reject(new Error(failureReason || "print failed"));
          }
        }
      );
    });

    printWindow.on("closed", () => {
      printWindow = null;
      resolve();
    });

    setTimeout(() => {
      cleanup();
      reject(new Error("print timeout"));
    }, 30000);
  });
});
