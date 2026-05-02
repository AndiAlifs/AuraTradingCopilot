import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { StrategyCardData } from '../../services/trade.service';

@Component({
  selector: 'app-strategy-card',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="bg-auraPanel border border-slate-700 rounded-xl overflow-hidden max-w-md w-full shadow-2xl mt-2 mb-4 relative group">
      <!-- Glow effect -->
      <div class="absolute -inset-0.5 bg-gradient-to-r from-auraGreen to-blue-500 rounded-xl blur opacity-25 group-hover:opacity-40 transition duration-1000"></div>
      
      <div class="relative bg-auraPanel rounded-xl p-5">
        <div class="flex justify-between items-start mb-4">
          <div>
            <h3 class="text-2xl font-black text-white tracking-tight">{{data.ticker}}</h3>
            <p class="text-xs text-slate-400 uppercase tracking-widest mt-1 font-semibold">Trade Setup</p>
          </div>
          <div class="text-right">
            <div class="text-3xl font-mono text-white">{{data.entry | number:'1.0-2'}}</div>
            <p class="text-xs text-slate-400 uppercase mt-1">Entry Price</p>
          </div>
        </div>

        <div class="mb-5">
          <div class="flex justify-between text-xs mb-1">
            <span class="text-slate-400">Aura Confidence</span>
            <span class="text-auraGreen font-bold">{{data.confidence}}%</span>
          </div>
          <div class="w-full bg-slate-900 rounded-full h-2">
            <div class="bg-gradient-to-r from-green-600 to-auraGreen h-2 rounded-full" [style.width.%]="data.confidence"></div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3 mb-5">
          <div class="bg-slate-900/50 border border-slate-700/50 p-3 rounded-lg">
            <div class="text-xs text-slate-400 uppercase mb-1">Take Profit</div>
            <div class="text-lg font-mono text-auraGreen font-bold">{{data.takeProfit | number:'1.0-2'}}</div>
          </div>
          <div class="bg-slate-900/50 border border-slate-700/50 p-3 rounded-lg">
            <div class="text-xs text-slate-400 uppercase mb-1">Stop Loss</div>
            <div class="text-lg font-mono text-auraRed font-bold">{{data.stopLoss | number:'1.0-2'}}</div>
          </div>
        </div>

        <div class="bg-slate-900/50 p-3 rounded-lg mb-5 border-l-2 border-auraGreen text-sm text-slate-300 leading-relaxed italic">
          "{{data.rationale}}"
        </div>

        <button (click)="setAlert()" class="w-full py-3 bg-slate-800 hover:bg-slate-700 border border-slate-600 text-white rounded-lg font-semibold transition-all hover:shadow-lg hover:border-auraGreen focus:outline-none focus:ring-2 focus:ring-auraGreen/50 flex justify-center items-center gap-2">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-auraGreen" viewBox="0 0 20 20" fill="currentColor">
            <path d="M10 2a6 6 0 00-6 6v3.586l-.707.707A1 1 0 004 14h12a1 1 0 00.707-1.707L16 11.586V8a6 6 0 00-6-6zM10 18a3 3 0 01-3-3h6a3 3 0 01-3 3z" />
          </svg>
          Set Alert
        </button>

        <!-- Toast -->
        <div *ngIf="showToast" class="absolute bottom-4 left-1/2 -translate-x-1/2 bg-auraGreen text-slate-900 px-4 py-2 rounded-md font-bold text-sm shadow-lg whitespace-nowrap animate-bounce">
          ✅ Alert Set: Aura is monitoring {{data.ticker}}
        </div>
      </div>
    </div>
  `
})
export class StrategyCardComponent {
  @Input() data!: StrategyCardData;
  showToast = false;

  setAlert() {
    this.showToast = true;
    setTimeout(() => this.showToast = false, 3000);
  }
}
