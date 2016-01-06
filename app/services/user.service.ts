import {Injectable} from 'angular2/core';
import {Response} from 'angular2/http';
import 'rxjs/add/operator/map';
import {AuthHttp} from 'angular2-jwt/angular2-jwt';


@Injectable()
export class UserService {

  constructor(private authHttp:AuthHttp) {  }

  getUser() {
    return this.authHttp.get('/api/user').map((res: Response) => res.json());
  }
}
