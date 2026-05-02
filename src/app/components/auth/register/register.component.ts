import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AbstractControl, FormBuilder, FormGroup, ReactiveFormsModule, ValidationErrors, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../services/auth.service';

function passwordsMatch(g: AbstractControl): ValidationErrors | null {
  const pw = g.get('password')?.value;
  const confirm = g.get('confirm')?.value;
  return pw === confirm ? null : { mismatch: true };
}

@Component({
  selector: 'app-register',
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
          <p class="text-slate-400 text-sm mt-2">Create your account</p>
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
            <input formControlName="password" type="password" placeholder="Min. 8 characters"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-3 text-white
                     placeholder-slate-500 focus:outline-none focus:border-auraNeon transition-colors text-sm" />
          </div>

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">Confirm Password</label>
            <input formControlName="confirm" type="password" placeholder="Repeat password"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-3 text-white
                     placeholder-slate-500 focus:outline-none focus:border-auraNeon transition-colors text-sm" />
            <p *ngIf="form.hasError('mismatch') && form.get('confirm')?.touched"
              class="text-red-400 text-xs mt-1">Passwords do not match</p>
          </div>

          <div *ngIf="error"
            class="bg-red-950/50 border border-red-800 text-red-400 rounded-lg px-4 py-3 text-sm">
            {{ error }}
          </div>

          <button type="submit" [disabled]="loading || form.invalid"
            class="w-full py-3 bg-auraNeon hover:bg-cyan-400 text-slate-900 font-bold rounded-lg
                   transition-colors disabled:opacity-40 disabled:cursor-not-allowed text-sm">
            {{ loading ? 'Creating account…' : 'Create Account' }}
          </button>

          <p class="text-center text-slate-400 text-sm">
            Already have an account?
            <a routerLink="/login" class="text-auraNeon hover:underline ml-1">Sign in</a>
          </p>
        </form>

        <!-- Footer Credit -->
        <div class="mt-8 text-center text-[10px] text-slate-600 space-y-1">
          <p>&copy; 2026 Aura Trading Copilot</p>
          <p>Made by <a href="https://andialifs.github.io/" target="_blank" class="text-auraNeon/60 hover:text-auraNeon transition-colors">Andi Alifsyah</a> (2025)</p>
        </div>
      </div>
    </div>
  `
})
export class RegisterComponent {
  form: FormGroup;
  loading = false;
  error = '';

  constructor(private fb: FormBuilder, private auth: AuthService, private router: Router) {
    this.form = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(8)]],
      confirm: ['', Validators.required]
    }, { validators: passwordsMatch });
    if (this.auth.isLoggedIn) this.router.navigate(['/dashboard']);
  }

  submit() {
    if (this.form.invalid) return;
    this.loading = true;
    this.error = '';
    const { email, password } = this.form.value;
    this.auth.register(email, password).subscribe({
      next: () => this.router.navigate(['/dashboard']),
      error: err => {
        this.error = typeof err.error === 'string' ? err.error : 'Registration failed. Try a different email.';
        this.loading = false;
      }
    });
  }
}
