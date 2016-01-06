import {Component, ViewEncapsulation} from 'angular2/core';
import {
  RouteConfig,
  ROUTER_DIRECTIVES,
  CanActivate
} from 'angular2/router';
import {AboutCmp} from '../about/about';
import {About2Cmp} from '../about2/about';
import {NameList} from '../../services/name_list';
import {UserIconCmp} from '../user/userIcon';
import {LogoutBtn} from '../user/logout';
import {Router} from 'angular2/router';
import {tokenNotExpired} from 'angular2-jwt/angular2-jwt';


@Component({
  selector: 'home',
  viewProviders: [NameList],
  templateUrl: './components/home/home.html',
  styleUrls: ['./components/home/home.css'],
  encapsulation: ViewEncapsulation.None,
  directives: [ROUTER_DIRECTIVES, UserIconCmp, LogoutBtn]
})
@CanActivate(()=>tokenNotExpired())
@RouteConfig([
  {path: '/', component: AboutCmp, as: 'Home'},
  {path: '/about2', component: About2Cmp, as: 'About2'},
])
export class HomeCmp {
  constructor(private router:Router) {
  }

}
