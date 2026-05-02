import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { TradeService, ScreenerResult, ScreenerFilters } from '../../services/trade.service';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-screener',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  template: `
    <div class="min-h-screen bg-slate-950 text-slate-200 font-sans">

      <!-- Nav -->
      <nav class="border-b border-auraBorder bg-auraPanel px-6 py-3 flex items-center justify-between">
        <a routerLink="/dashboard" class="text-xl font-black tracking-tight">
          <span class="text-white">Aura</span><span class="text-auraNeon"> Screener</span>
        </a>
        <div class="flex items-center gap-4 text-sm">
          <a routerLink="/dashboard" class="text-slate-400 hover:text-white transition-colors">Dashboard</a>
          <a routerLink="/alerts" class="text-slate-400 hover:text-white transition-colors">Alerts</a>
          <button (click)="logout()" class="text-slate-500 hover:text-slate-300 transition-colors">Sign out</button>
        </div>
      </nav>

      <div class="max-w-7xl mx-auto px-6 py-6">

        <!-- Filters -->
        <form [formGroup]="filters" (ngSubmit)="run()"
          class="bg-auraPanel border border-auraBorder rounded-xl p-5 mb-6 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">Min Price</label>
            <input formControlName="minPrice" type="number" placeholder="0"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-auraNeon" />
          </div>

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">Max Price</label>
            <input formControlName="maxPrice" type="number" placeholder="Any"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-auraNeon" />
          </div>

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">Min Volume</label>
            <input formControlName="minVolume" type="number" placeholder="0"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-auraNeon" />
          </div>

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">RSI Min</label>
            <input formControlName="minRSI" type="number" min="0" max="100" placeholder="0"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-auraNeon" />
          </div>

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">RSI Max</label>
            <input formControlName="maxRSI" type="number" min="0" max="100" placeholder="100"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-auraNeon" />
          </div>

          <div class="space-y-1">
            <label class="text-xs text-slate-400 uppercase tracking-wider">Signal</label>
            <select formControlName="signal"
              class="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-auraNeon">
              <option value="">All</option>
              <option value="bullish">Bullish</option>
              <option value="neutral">Neutral</option>
              <option value="bearish">Bearish</option>
            </select>
          </div>

          <div class="col-span-2 md:col-span-3 lg:col-span-6 flex gap-3 pt-1">
            <button type="submit" [disabled]="loading"
              class="px-6 py-2 bg-auraNeon hover:bg-cyan-400 text-slate-900 font-bold rounded-lg text-sm transition-colors disabled:opacity-50">
              {{ loading ? 'Scanning…' : 'Run Screener' }}
            </button>
            <button type="button" (click)="reset()"
              class="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-600 text-slate-300 font-semibold rounded-lg text-sm transition-colors">
              Reset
            </button>
          </div>
        </form>

        <!-- Error -->
        <div *ngIf="error" class="mb-4 bg-red-950/50 border border-red-800 text-red-400 rounded-lg px-4 py-3 text-sm">
          {{ error }}
        </div>

        <!-- Results table -->
        <div *ngIf="results !== null" class="bg-auraPanel border border-auraBorder rounded-xl overflow-hidden">
          <div class="px-5 py-4 border-b border-auraBorder flex items-center justify-between">
            <h2 class="font-bold text-white">
              {{ results.length }} {{ results.length === 1 ? 'match' : 'matches' }}
            </h2>
            <span class="text-xs text-slate-500">IDX top-20 universe</span>
          </div>

          <div *ngIf="results.length === 0" class="px-5 py-10 text-center text-slate-500 text-sm">
            No stocks match the current filters.
          </div>

          <table *ngIf="results.length > 0" class="w-full text-sm">
            <thead class="text-xs text-slate-400 uppercase tracking-wider border-b border-auraBorder">
              <tr>
                <th class="px-5 py-3 text-left">Ticker</th>
                <th class="px-4 py-3 text-right">Price</th>
                <th class="px-4 py-3 text-right">Change</th>
                <th class="px-4 py-3 text-right">Volume</th>
                <th class="px-4 py-3 text-right">RSI</th>
                <th class="px-4 py-3 text-right">MA20</th>
                <th class="px-4 py-3 text-right">MACD</th>
                <th class="px-4 py-3 text-center">Signal</th>
                <th class="px-4 py-3 text-center"></th>
              </tr>
            </thead>
            <tbody>
              <tr *ngFor="let r of results"
                class="border-b border-slate-800/50 hover:bg-slate-800/30 transition-colors">
                <td class="px-5 py-3 font-mono font-bold text-white">{{ r.ticker }}</td>
                <td class="px-4 py-3 text-right font-mono">{{ r.currentPrice | number:'1.0-0' }}</td>
                <td class="px-4 py-3 text-right font-mono"
                  [class.text-auraGreen]="r.percentChange >= 0"
                  [class.text-auraRed]="r.percentChange < 0">
                  {{ r.percentChange >= 0 ? '+' : '' }}{{ r.percentChange | number:'1.2-2' }}%
                </td>
                <td class="px-4 py-3 text-right text-slate-400 font-mono">{{ r.volume | number }}</td>
                <td class="px-4 py-3 text-right font-mono"
                  [class.text-auraGreen]="r.rsi < 40"
                  [class.text-auraRed]="r.rsi > 65"
                  [class.text-slate-300]="r.rsi >= 40 && r.rsi <= 65">
                  {{ r.rsi | number:'1.1-1' }}
                </td>
                <td class="px-4 py-3 text-right font-mono text-slate-400">{{ r.ma20 | number:'1.0-0' }}</td>
                <td class="px-4 py-3 text-right font-mono"
                  [class.text-auraGreen]="r.macd > r.macdSignal"
                  [class.text-auraRed]="r.macd <= r.macdSignal">
                  {{ r.macd | number:'1.2-2' }}
                </td>
                <td class="px-4 py-3 text-center">
                  <span class="px-2 py-0.5 rounded-full text-xs font-bold uppercase"
                    [class.bg-green-900]="r.signal === 'bullish'"
                    [class.text-auraGreen]="r.signal === 'bullish'"
                    [class.bg-red-900]="r.signal === 'bearish'"
                    [class.text-auraRed]="r.signal === 'bearish'"
                    [class.bg-slate-800]="r.signal === 'neutral'"
                    [class.text-slate-400]="r.signal === 'neutral'">
                    {{ r.signal }}
                  </span>
                </td>
                <td class="px-4 py-3 text-center">
                  <button (click)="analyze(r.ticker)"
                    class="px-3 py-1 bg-slate-800 hover:bg-slate-700 border border-slate-600
                           hover:border-auraNeon text-xs font-semibold rounded-lg transition-colors">
                    Analyze
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Footer Credit -->
        <div class="mt-12 py-8 border-t border-auraBorder/30 text-center text-xs text-slate-600 flex flex-col gap-2">
          <p>&copy; 2026 Aura Trading Copilot</p>
          <p>Made by <a href="https://andialifs.github.io/" target="_blank" class="text-auraNeon/60 hover:text-auraNeon transition-colors font-medium">Andi Alifsyah</a> (2025)</p>
        </div>
      </div>
    </div>
  `
})
export class ScreenerComponent implements OnInit {
  filters: FormGroup;
  results: ScreenerResult[] | null = null;
  loading = false;
  error = '';

  constructor(
    private fb: FormBuilder,
    private trade: TradeService,
    private auth: AuthService,
    private router: Router
  ) {
    this.filters = this.fb.group({
      minPrice: [null],
      maxPrice: [null],
      minVolume: [null],
      minRSI: [null],
      maxRSI: [null],
      signal: ['']
    });
  }

  ngOnInit() {
    this.run();
  }

  run() {
    this.loading = true;
    this.error = '';
    const v = this.filters.value;
    const f: ScreenerFilters = {};
    if (v.minPrice != null && v.minPrice !== '') f.minPrice = +v.minPrice;
    if (v.maxPrice != null && v.maxPrice !== '') f.maxPrice = +v.maxPrice;
    if (v.minVolume != null && v.minVolume !== '') f.minVolume = +v.minVolume;
    if (v.minRSI != null && v.minRSI !== '') f.minRSI = +v.minRSI;
    if (v.maxRSI != null && v.maxRSI !== '') f.maxRSI = +v.maxRSI;
    if (v.signal) f.signal = v.signal;

    this.trade.screen(f).subscribe({
      next: res => { this.results = res; this.loading = false; },
      error: () => { this.error = 'Screener unavailable. Check connection.'; this.loading = false; }
    });
  }

  reset() {
    this.filters.reset({ signal: '' });
    this.results = null;
  }

  analyze(ticker: string) {
    this.router.navigate(['/dashboard']).then(() => {
      // Small delay to ensure ChatComponent is initialized
      setTimeout(() => {
        this.trade.triggerChatWithTicker(ticker);
      }, 100);
    });
  }

  logout() { this.auth.logout(); }
}
