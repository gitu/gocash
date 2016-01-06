import {Component} from 'angular2/core';
import {AuthService} from '../../services/auth.service';
import {Router} from 'angular2/router';


@Component({
  selector: 'logout',
  viewProviders: [AuthService],
  template: '<a class="btn btn-danger btn-xs" role="button" (click)="logout()">logout</a>'
})
export class LogoutBtn {
  constructor(private authService: AuthService, private router:Router) {}
  logout() {
    this.authService.logout();
    this.router.navigateByUrl('/login');
  }
}
