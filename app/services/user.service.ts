import {Injectable} from 'angular2/core';
import {Http} from 'angular2/http';
import {Response} from 'angular2/http';
import 'rxjs/add/operator/map';


@Injectable()
export class UserService {

  constructor(private http:Http) {  }

  getUser() {
    return this.http.get('/api/user').map((res: Response) => res.json());
  }
}
