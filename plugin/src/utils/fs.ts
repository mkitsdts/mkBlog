import * as fs from 'fs/promises';
import * as path from 'path';

export interface UploadTask {
  mdPath: string;
  mdContent: string;
  images: { name: string; buffer: Buffer }[];
}

const IMG_EXT = new Set(['.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg']);

export async function findMarkdownFilesWithImageFolders(root: string): Promise<UploadTask[]> {
  const entries = await fs.readdir(root, { withFileTypes: true });
  const tasks: UploadTask[] = [];
  for (const e of entries) {
    if (e.isFile() && e.name.toLowerCase().endsWith('.md')) {
      const mdPath = path.join(root, e.name);
      const base = e.name.slice(0, -3); // remove .md
      const folder = path.join(root, base);
      const mdContent = await fs.readFile(mdPath, 'utf8');
      const images: { name: string; buffer: Buffer }[] = [];
      try {
        const imgEntries = await fs.readdir(folder, { withFileTypes: true });
        for (const ie of imgEntries) {
          if (ie.isFile() && IMG_EXT.has(path.extname(ie.name).toLowerCase())) {
            const buf = await fs.readFile(path.join(folder, ie.name));
            images.push({ name: ie.name, buffer: buf });
          }
        }
      } catch {
        // folder not exists -> ignore
      }
      tasks.push({ mdPath, mdContent, images });
    }
  }
  return tasks;
}
