import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, Observable, tap } from 'rxjs';
import { Router } from '@angular/router';

export interface AuthResponse {
  token: string;
  email: string;
}

const STORAGE_KEY = 'aura_auth';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private current = new BehaviorSubject<AuthResponse | null>(this.loadStored());
  currentUser$ = this.current.asObservable();

  constructor(private http: HttpClient, private router: Router) {}

  get isLoggedIn(): boolean {
    return this.current.value !== null;
  }

  get token(): string | null {
    return this.current.value?.token ?? null;
  }

  get email(): string | null {
    return this.current.value?.email ?? null;
  }

  register(email: string, password: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>('/api/register', { email, password }).pipe(
      tap(res => this.persist(res))
    );
  }

  login(email: string, password: string): Observable<AuthResponse> {
    return this.http.post<AuthResponse>('/api/login', { email, password }).pipe(
      tap(res => this.persist(res))
    );
  }

  updateProfile(profileData: { riskTolerance: string; preferredStrategy: string }): Observable<any> {
    return this.http.post('/api/profile', profileData);
  }

  logout() {
    localStorage.removeItem(STORAGE_KEY);
    this.current.next(null);
    this.router.navigate(['/login']);
  }

  private persist(res: AuthResponse) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(res));
    this.current.next(res);
  }

  private loadStored(): AuthResponse | null {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      return raw ? JSON.parse(raw) : null;
    } catch {
      return null;
    }
  }
}
