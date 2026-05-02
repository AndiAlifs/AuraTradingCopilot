import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../services/auth.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  template: `
    <div class="min-h-screen bg-slate-950 flex items-center justify-center p-4">
      <div class="w-full max-w-md">

        <div class="text-center mb-8">
          <div class="text-4xl font-black tracking-tight mb-1">
            <span class="text-white">Aura</span>
            <span class="text-auraNeon"> Trading</span>
          </div>
          <p class="text-slate-400 text-sm mt-2">Sign in to your account</p>
        </div>

        <form [formGroup]="form" (ngSubmit)="submit()"
          class="bg-auraPanel border border-auraBorder rounded-xl p-6 space-y-5">

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">Email</label>
            <input formControlName="email" type="email" placeholder="you@example.com"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-3 text-white
                     placeholder-slate-500 focus:outline-none focus:border-auraNeon transition-colors text-sm" />
          </div>

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">Password</label>
            <input formControlName="password" type="password" placeholder="••••••••"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-3 text-white
                     placeholder-slate-500 focus:outline-none focus:border-auraNeon transition-colors text-sm" />
          </div>

          <div *ngIf="error"
            class="bg-red-950/50 border border-red-800 text-red-400 rounded-lg px-4 py-3 text-sm">
            {{ error }}
          </div>

          <button type="submit" [disabled]="loading || form.invalid"
            class="w-full py-3 bg-auraNeon hover:bg-cyan-400 text-slate-900 font-bold rounded-lg
                   transition-colors disabled:opacity-40 disabled:cursor-not-allowed text-sm">
            {{ loading ? 'Signing in…' : 'Sign In' }}
          </button>

          <p class="text-center text-slate-400 text-sm">
            No account?
            <a routerLink="/register" class="text-auraNeon hover:underline ml-1">Sign up free</a>
          </p>
        </form>
      </div>
    </div>
  `
})
export class LoginComponent {
  form: FormGroup;
  loading = false;
  error = '';

  constructor(private fb: FormBuilder, private auth: AuthService, private router: Router) {
    this.form = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(8)]]
    });
    if (this.auth.isLoggedIn) this.router.navigate(['/dashboard']);
  }

  submit() {
    if (this.form.invalid) return;
    this.loading = true;
    this.error = '';
    const { email, password } = this.form.value;
    this.auth.login(email, password).subscribe({
      next: () => this.router.navigate(['/dashboard']),
      error: err => {
        this.error = typeof err.error === 'string' ? err.error : 'Invalid email or password';
        this.loading = false;
      }
    });
  }
}
