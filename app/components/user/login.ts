import {Component, ViewEncapsulation} from 'angular2/core';
import {AuthService} from '../../services/auth.service';
import {FormBuilder} from 'angular2/common';
import {Validators} from 'angular2/common';
import {Router} from 'angular2/router';


@Component({
  selector: 'login-page',
  viewProviders: [AuthService],
  templateUrl: './components/user/login.html',
  styleUrls: ['./components/user/login.css'],
  encapsulation: ViewEncapsulation.None
})
export class LoginCmp {
  loginForm;
  constructor(fb: FormBuilder, private authService: AuthService, private router:Router) {
    this.loginForm = fb.group({
      user: ['', Validators.required],
      password: ['', Validators.required]
    });
  }
  doLogin(event) {
    event.preventDefault();
    console.log(this.loginForm.value);
    this.authService.login(this.loginForm.value.user, this.loginForm.value.password).subscribe(
      (success)=>this.router.navigateByUrl('/'),
      (error)=>console.log(error)
    );
  }
}
