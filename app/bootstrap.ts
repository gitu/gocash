import {provide} from 'angular2/core';
import {bootstrap} from 'angular2/platform/browser';
import {ROUTER_PROVIDERS, LocationStrategy, HashLocationStrategy} from 'angular2/router';
import {AppCmp} from './components/app/app';
import {AuthConfig} from 'angular2-jwt/angular2-jwt';
import {AuthHttp} from 'angular2-jwt/angular2-jwt';
import {HTTP_PROVIDERS} from 'angular2/http';

bootstrap(AppCmp, [
  ROUTER_PROVIDERS,
  provide(LocationStrategy, { useClass: HashLocationStrategy }),
  provide(AuthConfig, {
    useFactory: () => {
      return new AuthConfig();
    }
  }),
  AuthHttp,
  HTTP_PROVIDERS
]);
