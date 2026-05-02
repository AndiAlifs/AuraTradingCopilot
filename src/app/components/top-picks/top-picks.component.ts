import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TradeService, TopPick } from '../../services/trade.service';

@Component({
  selector: 'app-top-picks',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="w-full bg-auraPanel border-b border-slate-700 p-2 flex overflow-x-auto items-center space-x-3 shadow-lg h-[60px]">
      <div class="text-xs font-semibold text-slate-500 uppercase tracking-widest pl-2 whitespace-nowrap">Live</div>
      <div *ngIf="loading" class="text-xs text-slate-500 italic px-2">Loading…</div>
      <div class="flex space-x-2 items-center">
        <button *ngFor="let pick of picks"
                (click)="onPick(pick)"
                class="flex items-center gap-2 bg-slate-900 px-3 py-1.5 rounded-md hover:bg-slate-800
                       transition-colors border border-slate-700/50 cursor-pointer whitespace-nowrap group">
          <span class="font-bold text-slate-200 text-sm">{{ pick.ticker | slice:0:-3 }}</span>
          <span class="text-xs font-mono" [class.text-auraGreen]="pick.percentChange >= 0" [class.text-auraRed]="pick.percentChange < 0">
            {{ pick.currentPrice | number:'1.0-0' }}
            <span>({{ pick.percentChange >= 0 ? '+' : '' }}{{ pick.percentChange | number:'1.1-1' }}%)</span>
          </span>
          <span *ngIf="pick.signal" class="text-xs px-1.5 py-0.5 rounded font-bold uppercase hidden group-hover:inline-block"
            [class.bg-green-900]="pick.signal === 'bullish'"
            [class.text-auraGreen]="pick.signal === 'bullish'"
            [class.bg-red-900]="pick.signal === 'bearish'"
            [class.text-auraRed]="pick.signal === 'bearish'"
            [class.bg-slate-800]="pick.signal === 'neutral'"
            [class.text-slate-400]="pick.signal === 'neutral'">
            {{ pick.signal }}
          </span>
        </button>
      </div>
    </div>
  `
})
export class TopPicksComponent implements OnInit {
  picks: TopPick[] = [];
  loading = false;

  constructor(private tradeService: TradeService) {}

  ngOnInit() {
    this.loading = true;
    this.tradeService.getTopPicks().subscribe({
      next: (data) => { this.picks = data; this.loading = false; },
      error: () => { this.loading = false; }
    });
  }

  onPick(pick: TopPick) {
    this.tradeService.triggerChatWithTicker(pick.ticker);
  }
}
