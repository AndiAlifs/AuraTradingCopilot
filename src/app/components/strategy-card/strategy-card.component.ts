import { Component, Input, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StrategyCardData, TradeService } from '../../services/trade.service';
import { AuthService } from '../../services/auth.service';

type Tab = 'trade' | 'ta' | 'risk';

@Component({
  selector: 'app-strategy-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="relative group max-w-md w-full mt-2 mb-4">
      <!-- Glow -->
      <div class="absolute -inset-0.5 bg-gradient-to-r from-auraGreen to-blue-500 rounded-xl blur opacity-20 group-hover:opacity-35 transition duration-1000 pointer-events-none"></div>

      <div class="relative bg-auraPanel border border-slate-700 rounded-xl overflow-hidden">

        <!-- Header -->
        <div class="px-5 pt-5 pb-3 flex justify-between items-start">
          <div>
            <h3 class="text-2xl font-black text-white tracking-tight">{{ data.ticker }}</h3>
            <p class="text-xs text-slate-400 uppercase tracking-widest mt-0.5 font-semibold">Aura Trade Setup</p>
          </div>
          <div class="text-right">
            <div class="text-3xl font-mono text-white leading-tight">{{ data.entry | number:'1.0-0' }}</div>
            <p class="text-xs text-slate-400 uppercase mt-0.5">Entry</p>
          </div>
        </div>

        <!-- Confidence bar -->
        <div class="px-5 pb-4">
          <div class="flex justify-between text-xs mb-1">
            <span class="text-slate-400">Aura Confidence</span>
            <span class="text-auraGreen font-bold">{{ data.confidence }}%</span>
          </div>
          <div class="w-full bg-slate-900 rounded-full h-1.5">
            <div class="bg-gradient-to-r from-green-600 to-auraGreen h-1.5 rounded-full transition-all"
              [style.width.%]="data.confidence"></div>
          </div>
        </div>

        <!-- Tab bar -->
        <div class="flex border-t border-b border-slate-700/50">
          <button *ngFor="let t of tabs"
            (click)="activeTab = t.key"
            class="flex-1 py-2.5 text-xs font-bold uppercase tracking-wider transition-colors"
            [class.text-auraNeon]="activeTab === t.key"
            [class.border-b-2]="activeTab === t.key"
            [class.border-auraNeon]="activeTab === t.key"
            [class.text-slate-500]="activeTab !== t.key"
            [class.hover:text-slate-300]="activeTab !== t.key">
            {{ t.label }}
          </button>
        </div>

        <!-- Tab: Trade Setup -->
        <div *ngIf="activeTab === 'trade'" class="p-5 space-y-4">
          <div class="grid grid-cols-2 gap-3">
            <div class="bg-slate-900/60 border border-slate-700/50 p-3 rounded-lg">
              <div class="text-xs text-slate-400 uppercase mb-1">Take Profit</div>
              <div class="text-lg font-mono text-auraGreen font-bold">{{ data.takeProfit | number:'1.0-0' }}</div>
              <div *ngIf="data.takeProfitJustification"
                class="text-xs text-slate-500 mt-1 leading-snug">
                {{ data.takeProfitJustification }}
              </div>
            </div>
            <div class="bg-slate-900/60 border border-slate-700/50 p-3 rounded-lg">
              <div class="text-xs text-slate-400 uppercase mb-1">Stop Loss</div>
              <div class="text-lg font-mono text-auraRed font-bold">{{ data.stopLoss | number:'1.0-0' }}</div>
              <div *ngIf="data.stopLossJustification"
                class="text-xs text-slate-500 mt-1 leading-snug">
                {{ data.stopLossJustification }}
              </div>
            </div>
          </div>

          <div class="bg-slate-900/50 border-l-2 border-auraGreen p-3 rounded-lg text-sm text-slate-300 leading-relaxed italic">
            "{{ data.rationale }}"
          </div>

          <div class="grid grid-cols-3 gap-2 text-center text-xs">
            <div class="bg-slate-900/40 rounded-lg p-2">
              <div class="text-slate-400 mb-0.5">R:R Ratio</div>
              <div class="font-mono font-bold text-white">{{ rrRatio }}</div>
            </div>
            <div class="bg-slate-900/40 rounded-lg p-2">
              <div class="text-slate-400 mb-0.5">Upside</div>
              <div class="font-mono font-bold text-auraGreen">+{{ upside }}%</div>
            </div>
            <div class="bg-slate-900/40 rounded-lg p-2">
              <div class="text-slate-400 mb-0.5">Downside</div>
              <div class="font-mono font-bold text-auraRed">-{{ downside }}%</div>
            </div>
          </div>
        </div>

        <!-- Tab: Technical Analysis -->
        <div *ngIf="activeTab === 'ta'" class="p-5">
          <div *ngIf="hasTa; else noTa">
            <div class="grid grid-cols-2 gap-3 mb-4">
              <!-- RSI -->
              <div class="bg-slate-900/60 border border-slate-700/50 p-3 rounded-lg">
                <div class="flex justify-between items-center mb-2">
                  <span class="text-xs text-slate-400 uppercase">RSI (14)</span>
                  <span class="text-xs font-bold px-2 py-0.5 rounded"
                    [class.text-auraGreen]="data.rsi! < 40"
                    [class.bg-green-900]="data.rsi! < 40"
                    [class.text-yellow-400]="data.rsi! >= 40 && data.rsi! <= 60"
                    [class.bg-yellow-900]="data.rsi! >= 40 && data.rsi! <= 60"
                    [class.text-auraRed]="data.rsi! > 60"
                    [class.bg-red-900]="data.rsi! > 60">
                    {{ rsiLabel }}
                  </span>
                </div>
                <div class="text-2xl font-mono font-bold text-white">{{ data.rsi | number:'1.1-1' }}</div>
                <div class="w-full bg-slate-800 rounded-full h-1 mt-2">
                  <div class="h-1 rounded-full transition-all"
                    [class.bg-auraGreen]="data.rsi! < 40"
                    [class.bg-yellow-400]="data.rsi! >= 40 && data.rsi! <= 60"
                    [class.bg-auraRed]="data.rsi! > 60"
                    [style.width.%]="data.rsi"></div>
                </div>
              </div>

              <!-- MACD -->
              <div class="bg-slate-900/60 border border-slate-700/50 p-3 rounded-lg">
                <div class="text-xs text-slate-400 uppercase mb-2">MACD (12/26/9)</div>
                <div class="text-2xl font-mono font-bold"
                  [class.text-auraGreen]="(data.macd ?? 0) > (data.macdSignal ?? 0)"
                  [class.text-auraRed]="(data.macd ?? 0) <= (data.macdSignal ?? 0)">
                  {{ data.macd | number:'1.2-2' }}
                </div>
                <div class="text-xs text-slate-400 mt-1">Signal: {{ data.macdSignal | number:'1.2-2' }}</div>
                <div class="text-xs mt-0.5"
                  [class.text-auraGreen]="(data.macd ?? 0) > (data.macdSignal ?? 0)"
                  [class.text-auraRed]="(data.macd ?? 0) <= (data.macdSignal ?? 0)">
                  {{ (data.macd ?? 0) > (data.macdSignal ?? 0) ? 'Bullish crossover' : 'Bearish crossover' }}
                </div>
              </div>
            </div>

            <!-- Moving Averages -->
            <div class="bg-slate-900/60 border border-slate-700/50 p-3 rounded-lg mb-3">
              <div class="text-xs text-slate-400 uppercase mb-3">Moving Averages</div>
              <div class="grid grid-cols-2 gap-4">
                <div *ngIf="data.ma20">
                  <div class="text-xs text-slate-500 mb-0.5">MA 20</div>
                  <div class="font-mono font-bold text-white">{{ data.ma20 | number:'1.0-0' }}</div>
                  <div class="text-xs mt-0.5"
                    [class.text-auraGreen]="data.entry >= data.ma20"
                    [class.text-auraRed]="data.entry < data.ma20">
                    Price {{ data.entry >= data.ma20 ? 'above' : 'below' }} MA20
                  </div>
                </div>
                <div *ngIf="data.ma50">
                  <div class="text-xs text-slate-500 mb-0.5">MA 50</div>
                  <div class="font-mono font-bold text-white">{{ data.ma50 | number:'1.0-0' }}</div>
                  <div class="text-xs mt-0.5"
                    [class.text-auraGreen]="data.entry >= data.ma50"
                    [class.text-auraRed]="data.entry < data.ma50">
                    Price {{ data.entry >= data.ma50 ? 'above' : 'below' }} MA50
                  </div>
                </div>
              </div>
            </div>

            <!-- ATR -->
            <div *ngIf="data.atr" class="bg-slate-900/60 border border-slate-700/50 p-3 rounded-lg">
              <div class="flex justify-between">
                <span class="text-xs text-slate-400 uppercase">ATR (14)</span>
                <span class="font-mono font-bold text-white">{{ data.atr | number:'1.0-0' }}</span>
              </div>
              <div class="text-xs text-slate-500 mt-1">Average volatility per day</div>
            </div>
          </div>

          <ng-template #noTa>
            <div class="py-8 text-center text-slate-500 text-sm">
              Technical indicators are calculated by the <code class="text-auraNeon">/api/analyze</code> endpoint.<br>
              <span class="text-xs">Chat-generated setups show trade levels only.</span>
            </div>
          </ng-template>
        </div>

        <!-- Tab: Risk -->
        <div *ngIf="activeTab === 'risk'" class="p-5 space-y-3">
          <div class="bg-slate-900/60 border border-slate-700/50 p-4 rounded-lg space-y-3">
            <div class="flex justify-between text-sm">
              <span class="text-slate-400">Entry Price</span>
              <span class="font-mono font-bold text-white">{{ data.entry | number:'1.0-0' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-slate-400">Take Profit</span>
              <span class="font-mono font-bold text-auraGreen">{{ data.takeProfit | number:'1.0-0' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-slate-400">Stop Loss</span>
              <span class="font-mono font-bold text-auraRed">{{ data.stopLoss | number:'1.0-0' }}</span>
            </div>
            <div class="border-t border-slate-700 pt-3 flex justify-between text-sm">
              <span class="text-slate-400">Reward / Risk</span>
              <span class="font-mono font-bold text-white">{{ rrRatio }}</span>
            </div>
          </div>

          <div *ngIf="data.stopLossJustification"
            class="bg-slate-900/50 border-l-2 border-auraRed p-3 rounded-lg">
            <div class="text-xs text-slate-400 uppercase mb-1">Stop Loss Basis</div>
            <p class="text-sm text-slate-300">{{ data.stopLossJustification }}</p>
          </div>

          <div *ngIf="data.takeProfitJustification"
            class="bg-slate-900/50 border-l-2 border-auraGreen p-3 rounded-lg">
            <div class="text-xs text-slate-400 uppercase mb-1">Target Basis</div>
            <p class="text-sm text-slate-300">{{ data.takeProfitJustification }}</p>
          </div>

          <div class="bg-amber-950/30 border border-amber-800/50 p-3 rounded-lg text-xs text-amber-400">
            Risk max 1–2% of account per trade. Size your position so that hitting the stop loss equals your max risk.
          </div>
        </div>

        <!-- Set Alert button -->
        <div class="px-5 pb-5">
          <button (click)="setAlert()"
            class="w-full py-2.5 bg-slate-800 hover:bg-slate-700 border border-slate-600
                   hover:border-auraGreen text-white rounded-lg font-semibold text-sm transition-all
                   focus:outline-none focus:ring-2 focus:ring-auraGreen/40 flex justify-center items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-auraGreen" viewBox="0 0 20 20" fill="currentColor">
              <path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z" />
            </svg>
            Set Price Alert
          </button>
        </div>

        <!-- Toast -->
        <div *ngIf="showToast"
          class="absolute bottom-4 left-1/2 -translate-x-1/2 bg-auraGreen text-slate-900 px-4 py-2 rounded-md font-bold text-sm shadow-lg whitespace-nowrap animate-bounce">
          Alert set for {{ data.ticker }}
        </div>
      </div>
    </div>
  `
})
export class StrategyCardComponent implements OnInit {
  @Input() data!: StrategyCardData;

  activeTab: Tab = 'trade';
  showToast = false;

  tabs = [
    { key: 'trade' as Tab, label: 'Trade Setup' },
    { key: 'ta' as Tab, label: 'Technical' },
    { key: 'risk' as Tab, label: 'Risk' }
  ];

  constructor(private trade: TradeService, private auth: AuthService) {}

  ngOnInit() {}

  get hasTa(): boolean {
    return this.data.rsi != null;
  }

  get rsiLabel(): string {
    if (!this.data.rsi) return '';
    if (this.data.rsi < 30) return 'Oversold';
    if (this.data.rsi < 50) return 'Weak';
    if (this.data.rsi <= 60) return 'Neutral';
    if (this.data.rsi <= 70) return 'Strong';
    return 'Overbought';
  }

  get upside(): string {
    return (((this.data.takeProfit - this.data.entry) / this.data.entry) * 100).toFixed(1);
  }

  get downside(): string {
    return (((this.data.entry - this.data.stopLoss) / this.data.entry) * 100).toFixed(1);
  }

  get rrRatio(): string {
    const reward = this.data.takeProfit - this.data.entry;
    const risk = this.data.entry - this.data.stopLoss;
    if (risk <= 0) return '—';
    return (reward / risk).toFixed(1) + ':1';
  }

  setAlert() {
    if (this.auth.isLoggedIn) {
      this.trade.createAlert({
        ticker: this.data.ticker,
        condition: 'above',
        targetPrice: this.data.takeProfit
      }).subscribe();
    }
    this.showToast = true;
    setTimeout(() => this.showToast = false, 3000);
  }
}
