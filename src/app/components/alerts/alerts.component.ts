import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { TradeService, Alert } from '../../services/trade.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-alerts',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  template: `
    <div class="min-h-screen bg-slate-950 text-slate-200 font-sans">

      <!-- Nav -->
      <nav class="border-b border-auraBorder bg-auraPanel px-6 py-3 flex items-center justify-between">
        <a routerLink="/dashboard" class="text-xl font-black tracking-tight">
          <span class="text-white">Aura</span><span class="text-auraNeon"> Alerts</span>
        </a>
        <div class="flex items-center gap-4 text-sm">
          <a routerLink="/dashboard" class="text-slate-400 hover:text-white transition-colors">Dashboard</a>
          <a routerLink="/screener" class="text-slate-400 hover:text-white transition-colors">Screener</a>
          <button (click)="logout()" class="text-slate-500 hover:text-slate-300 transition-colors">Sign out</button>
        </div>
      </nav>

      <div class="max-w-3xl mx-auto px-6 py-6 space-y-6">

        <!-- Create Alert Form -->
        <div class="bg-auraPanel border border-auraBorder rounded-xl p-5">
          <h2 class="font-bold text-white mb-4">New Price Alert</h2>
          <form [formGroup]="form" (ngSubmit)="create()" class="grid grid-cols-1 md:grid-cols-3 gap-4">

            <div class="space-y-1">
              <label class="text-xs text-slate-400 uppercase tracking-wider">Ticker</label>
              <input formControlName="ticker" type="text" placeholder="e.g. BBCA"
                class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm
                       focus:outline-none focus:border-auraNeon transition-colors uppercase" />
            </div>

            <div class="space-y-1">
              <label class="text-xs text-slate-400 uppercase tracking-wider">Condition</label>
              <select formControlName="condition"
                class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm
                       focus:outline-none focus:border-auraNeon transition-colors">
                <option value="above">Price goes above</option>
                <option value="below">Price goes below</option>
              </select>
            </div>

            <div class="space-y-1">
              <label class="text-xs text-slate-400 uppercase tracking-wider">Target Price (IDR)</label>
              <input formControlName="targetPrice" type="number" placeholder="e.g. 10000"
                class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm
                       focus:outline-none focus:border-auraNeon transition-colors" />
            </div>

            <div class="md:col-span-3 flex items-center gap-3">
              <button type="submit" [disabled]="createLoading || form.invalid"
                class="px-6 py-2 bg-auraNeon hover:bg-cyan-400 text-slate-900 font-bold rounded-lg text-sm
                       transition-colors disabled:opacity-40 disabled:cursor-not-allowed">
                {{ createLoading ? 'Setting…' : 'Set Alert' }}
              </button>
              <span *ngIf="createSuccess" class="text-auraGreen text-sm font-semibold">Alert created!</span>
              <span *ngIf="createError" class="text-auraRed text-sm">{{ createError }}</span>
            </div>
          </form>
        </div>

        <!-- Active Alerts List -->
        <div class="bg-auraPanel border border-auraBorder rounded-xl overflow-hidden">
          <div class="px-5 py-4 border-b border-auraBorder flex items-center justify-between">
            <h2 class="font-bold text-white">Active Alerts</h2>
            <span class="text-xs text-slate-500">Checked every 5 minutes</span>
          </div>

          <div *ngIf="loading" class="px-5 py-10 text-center text-slate-500 text-sm">
            Loading alerts…
          </div>

          <div *ngIf="!loading && alerts.length === 0"
            class="px-5 py-10 text-center text-slate-500 text-sm">
            No active alerts. Set one above to get notified when a price target is hit.
          </div>

          <ul *ngIf="!loading && alerts.length > 0" class="divide-y divide-slate-800">
            <li *ngFor="let a of alerts" class="px-5 py-4 flex items-center justify-between">
              <div class="flex items-center gap-4">
                <div class="w-10 h-10 bg-slate-900 rounded-lg flex items-center justify-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-auraNeon" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z" />
                  </svg>
                </div>
                <div>
                  <div class="font-mono font-bold text-white">{{ a.ticker }}</div>
                  <div class="text-sm text-slate-400">
                    Alert when price
                    <span [class.text-auraGreen]="a.condition === 'above'"
                          [class.text-auraRed]="a.condition === 'below'" class="font-semibold">
                      {{ a.condition }}
                    </span>
                    <span class="font-mono text-white ml-1">{{ a.targetPrice | number:'1.0-0' }}</span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-3">
                <span class="text-xs text-slate-500">{{ a.createdAt | date:'dd MMM' }}</span>
                <button (click)="remove(a.id)"
                  class="p-1.5 text-slate-500 hover:text-auraRed transition-colors rounded">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </li>
          </ul>
        </div>

      </div>
    </div>
  `
})
export class AlertsComponent implements OnInit {
  form: FormGroup;
  alerts: Alert[] = [];
  loading = false;
  createLoading = false;
  createSuccess = false;
  createError = '';

  constructor(
    private fb: FormBuilder,
    private trade: TradeService,
    private auth: AuthService
  ) {
    this.form = this.fb.group({
      ticker: ['', [Validators.required, Validators.pattern(/^[A-Za-z]+$/)]],
      condition: ['above', Validators.required],
      targetPrice: [null, [Validators.required, Validators.min(1)]]
    });
  }

  ngOnInit() { this.load(); }

  load() {
    this.loading = true;
    this.trade.getAlerts().subscribe({
      next: data => { this.alerts = data; this.loading = false; },
      error: () => { this.loading = false; }
    });
  }

  create() {
    if (this.form.invalid) return;
    this.createLoading = true;
    this.createError = '';
    this.createSuccess = false;

    const { ticker, condition, targetPrice } = this.form.value;
    this.trade.createAlert({ ticker, condition, targetPrice: +targetPrice }).subscribe({
      next: () => {
        this.createSuccess = true;
        this.createLoading = false;
        this.form.reset({ condition: 'above' });
        this.trade.notifyAlertsChanged();
        this.load();
        setTimeout(() => this.createSuccess = false, 3000);
      },
      error: err => {
        this.createError = typeof err.error === 'string' ? err.error : 'Failed to create alert';
        this.createLoading = false;
      }
    });
  }

  remove(id: number) {
    this.trade.deleteAlert(id).subscribe({
      next: () => {
        this.alerts = this.alerts.filter(a => a.id !== id);
        this.trade.notifyAlertsChanged();
      },
      error: () => {}
    });
  }

  logout() { this.auth.logout(); }
}
