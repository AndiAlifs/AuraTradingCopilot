import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TradeService, TopPick } from '../../services/trade.service';

@Component({
  selector: 'app-top-picks',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="w-full bg-auraPanel border-b border-slate-700 p-2 flex overflow-x-auto items-center space-x-4 shadow-lg h-[60px]">
      <div class="text-sm font-semibold text-slate-400 uppercase tracking-widest pl-2 whitespace-nowrap">Top Picks</div>
      <div class="flex-1 flex space-x-3 items-center">
        <button *ngFor="let pick of picks" 
                (click)="onPick(pick)"
                class="flex items-center space-x-2 bg-slate-900 px-3 py-1.5 rounded-md hover:bg-slate-700 transition-colors border border-slate-700/50 cursor-pointer whitespace-nowrap">
          <span class="font-bold text-slate-200">{{pick.ticker}}</span>
          <span class="text-xs font-mono" [class.text-auraGreen]="pick.percentChange >= 0" [class.text-auraRed]="pick.percentChange < 0">
            {{pick.currentPrice}} ({{pick.percentChange > 0 ? '+' : ''}}{{pick.percentChange}}%)
          </span>
        </button>
      </div>
    </div>
  `
})
export class TopPicksComponent implements OnInit {
  picks: TopPick[] = [];

  constructor(private tradeService: TradeService) {}

  ngOnInit() {
    this.tradeService.getTopPicks().subscribe({
      next: (data) => this.picks = data,
      error: (err) => console.error('Failed to load top picks', err)
    });
  }

  onPick(pick: TopPick) {
    this.tradeService.triggerChatWithTicker(pick.ticker);
  }
}
