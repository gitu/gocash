import {Component} from 'angular2/core';
import {UserService} from '../../services/user.service';

@Component({
  selector: 'user-icon',
  viewProviders: [UserService],
  templateUrl: './components/user/userIcon.html',
  styleUrls: ['./components/user/userIcon.css']
})
export class UserIconCmp {
  user;
  constructor(private userService: UserService) {  }

  ngOnInit() {
    this.userService.getUser()
      .subscribe((user:any) => {
        this.user = user;
      });
  }
}
