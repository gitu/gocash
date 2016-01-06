import {Component} from 'angular2/core';
import {UserService} from '../../services/user.service';
import {tokenNotExpired} from 'angular2-jwt/angular2-jwt';

@Component({
  selector: 'user-icon',
  viewProviders: [UserService],
  template: '<pre>{{ user }}</pre>'
})
export class UserIconCmp {
  user;
  constructor(private userService: UserService) {  }

  ngOnInit() {
    if (tokenNotExpired()) {
      this.userService.getUser()
        .subscribe((user:any) => {
          this.user = user;
        });
    }
  }
}
