import * as util from 'gulp-util';
import * as express from 'express';
import * as openResource from 'open';
import * as serveStatic from 'serve-static';
import * as codeChangeTool from './code_change_tools';
import {resolve} from 'path';
import {APP_BASE, APP_DEST, DOCS_DEST, DOCS_PORT, PORT, BACKEND_PORT} from '../config';

export function serveSPA() {
  let server = express();
  let proxy = require('rocky')();
  codeChangeTool.listen();


  proxy.get('/api/*').forward('http://localhost:' + BACKEND_PORT);
  proxy.put('/api/*').forward('http://localhost:' + BACKEND_PORT);
  proxy.post('/api/*').forward('http://localhost:' + BACKEND_PORT);
  proxy.delete('/api/*').forward('http://localhost:' + BACKEND_PORT);
  proxy.post('/auth').forward('http://localhost:' + BACKEND_PORT);


  proxy.get('/').redirect(APP_BASE + APP_DEST);

  server.use(proxy.middleware());

  server.use.apply(server, codeChangeTool.middleware);


  server.listen(PORT, () => {
    util.log('Server is listening on port: ' + PORT);
    openResource('http://localhost:' + PORT + APP_BASE + APP_DEST);
  });
}

export function notifyLiveReload(e) {
  let fileName = e.path;
  codeChangeTool.changed(fileName);
}

export function serveDocs() {
  let server = express();

  server.use(
    APP_BASE,
    serveStatic(resolve(process.cwd(), DOCS_DEST))
  );

  server.listen(DOCS_PORT, () =>
    openResource('http://localhost:' + DOCS_PORT + APP_BASE)
  );
}
