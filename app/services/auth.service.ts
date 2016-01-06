import {Injectable} from 'angular2/core';
import {Response} from 'angular2/http';
import 'rxjs/add/operator/map';
import {Http} from 'angular2/http';
import {Observable} from 'rxjs/Observable';


@Injectable()
export class AuthService {

  constructor(private http:Http) {
  }

  login(userid, password):Observable<String> {
    var observable = Observable.create(
      (observer) =>
        this.http
          .post('/auth', JSON.stringify({'userid': userid, 'password': password}))
          .map((res:Response) => res.json())
          .subscribe(
            (resp)=> {
              localStorage.setItem('id_token', resp['token']);
              observer.next('success');
            },
            (error)=>  {
              localStorage.removeItem('id_token');
              observer.error(error);
            },
            () => observer.complete()
          )
    );
    return observable;
  }

  logout() {
    localStorage.removeItem('id_token');
  }
}
