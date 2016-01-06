import {Component, ViewEncapsulation} from 'angular2/core';
import {
  RouteConfig,
  ROUTER_DIRECTIVES
} from 'angular2/router';
import {tokenNotExpired} from 'angular2-jwt/angular2-jwt';
import {Router} from 'angular2/router';

import {HomeCmp} from '../home/home';
import {LoginCmp} from '../user/login';

@Component({
  selector: 'app',
  templateUrl: './components/app/app.html',
  styleUrls: ['./components/app/app.css'],
  encapsulation: ViewEncapsulation.None,
  directives: [ROUTER_DIRECTIVES]
})
@RouteConfig([
  {path: '/m/...', component: HomeCmp, as: 'Main'},
  {path: '/login', component: LoginCmp, as: 'Login'},
  {path: '/',      redirectTo: ['/Main/Home']}
])
export class AppCmp {
  constructor(private router:Router) {
    if (!tokenNotExpired()) {
      this.router.navigateByUrl('/login');
    }
  }
}
