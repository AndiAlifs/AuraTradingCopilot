import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { TopPicksComponent } from '../top-picks/top-picks.component';
import { ChatComponent } from '../chat/chat.component';
import { TradingProfileModalComponent } from '../trading-profile-modal/trading-profile-modal.component';
import { AuthService } from '../../services/auth.service';
import { TradeService, ScreenerResult, Alert } from '../../services/trade.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink, TopPicksComponent, ChatComponent, TradingProfileModalComponent],
  template: `
    <div class="flex flex-col h-screen bg-slate-950 font-sans text-slate-200">

      <!-- Top nav -->
      <header class="flex items-center justify-between px-4 py-2 border-b border-auraBorder bg-auraPanel shrink-0">
        <div class="flex items-center gap-1 text-lg font-black tracking-tight">
          <span class="text-white">Aura</span>
          <span class="text-auraNeon"> Copilot</span>
        </div>
        <nav class="flex items-center gap-3 text-sm">
          <a routerLink="/screener"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L13 13.414V19a1 1 0 01-.553.894l-4 2A1 1 0 017 21v-7.586L3.293 6.707A1 1 0 013 6V4z" />
            </svg>
            Screener
          </a>
          <a routerLink="/alerts"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            Alerts
          </a>
          <ng-container *ngIf="email">
            <button (click)="showProfileModal = true"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              Profile
            </button>
            <div class="h-4 w-px bg-slate-700"></div>
            <span class="text-xs text-slate-500 hidden sm:block">{{ email }}</span>
            <button (click)="logout()"
              class="px-3 py-1.5 text-xs text-slate-500 hover:text-slate-300 transition-colors rounded-lg hover:bg-slate-800">
              Sign out
            </button>
          </ng-container>
          <ng-container *ngIf="!email">
            <div class="h-4 w-px bg-slate-700"></div>
            <a routerLink="/login"
              class="px-3 py-1.5 text-xs text-auraNeon hover:text-cyan-300 transition-colors rounded-lg hover:bg-slate-800">
              Sign in
            </a>
          </ng-container>
        </nav>
      </header>

      <!-- Ticker bar -->
      <app-top-picks></app-top-picks>

      <main class="flex-1 overflow-hidden flex bg-slate-950">
        <!-- Chat area -->
        <div class="flex-1 flex flex-col overflow-hidden">
          <app-chat></app-chat>
        </div>

        <!-- Right Panel (Market Pulse / Dashboard widgets) -->
        <aside class="w-80 lg:w-96 border-l border-auraBorder bg-auraPanel flex flex-col overflow-y-auto hidden md:flex shrink-0">
          <div class="p-5 space-y-6">
            
            <!-- Quick Stats -->
            <div>
              <h3 class="text-xs font-bold text-slate-500 uppercase tracking-widest mb-3">Market Pulse</h3>
              <div class="grid grid-cols-2 gap-3">
                <div class="bg-slate-900 border border-slate-800 rounded-lg p-3">
                  <div class="text-slate-500 text-xs mb-1">Active Alerts</div>
                  <div class="text-xl font-mono text-white">{{ alerts.length }}</div>
                </div>
                <div class="bg-slate-900 border border-slate-800 rounded-lg p-3">
                  <div class="text-slate-500 text-xs mb-1">Bullish Setups</div>
                  <div class="text-xl font-mono text-auraGreen">{{ bullishStocks.length }}</div>
                </div>
              </div>
            </div>

            <!-- Bullish Screener Preview -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-xs font-bold text-slate-500 uppercase tracking-widest">Bullish Movers</h3>
                <a routerLink="/screener" class="text-xs text-auraNeon hover:text-cyan-300">View All</a>
              </div>
              <div class="space-y-2">
                <div *ngIf="bullishStocks.length === 0 && !loading" class="text-sm text-slate-500 italic">No bullish signals found.</div>
                <div *ngIf="loading" class="text-sm text-slate-500 italic">Loading...</div>
                <div *ngFor="let s of bullishStocks" class="bg-slate-900 border border-slate-800 rounded-lg p-3 hover:border-slate-600 transition-colors cursor-pointer group" (click)="analyze(s.ticker)">
                  <div class="flex justify-between items-center mb-1">
                    <span class="font-mono font-bold text-white group-hover:text-auraNeon transition-colors">{{ s.ticker }}</span>
                    <span class="font-mono text-sm" [class.text-auraGreen]="s.percentChange >= 0" [class.text-auraRed]="s.percentChange < 0">
                      {{ s.percentChange >= 0 ? '+' : '' }}{{ s.percentChange | number:'1.2-2' }}%
                    </span>
                  </div>
                  <div class="flex justify-between items-center text-xs text-slate-400">
                    <span>{{ s.currentPrice | number:'1.0-0' }} IDR</span>
                    <span>Vol: {{ (s.volume / 1000000) | number:'1.1-1' }}M</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Recent Alerts -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-xs font-bold text-slate-500 uppercase tracking-widest">Your Alerts</h3>
                <a routerLink="/alerts" class="text-xs text-auraNeon hover:text-cyan-300">Manage</a>
              </div>
              <div class="space-y-2">
                <div *ngIf="alerts.length === 0 && !loadingAlerts" class="text-sm text-slate-500 italic">No active alerts.</div>
                <div *ngIf="loadingAlerts" class="text-sm text-slate-500 italic">Loading...</div>
                <div *ngFor="let a of alerts" class="bg-slate-900 border border-slate-800 rounded-lg p-3">
                  <div class="flex justify-between items-center mb-1">
                    <span class="font-mono font-bold text-white">{{ a.ticker }}</span>
                    <span class="text-xs text-slate-500">{{ a.createdAt | date:'MMM d' }}</span>
                  </div>
                  <div class="text-xs text-slate-400">
                    <span [class.text-auraGreen]="a.condition === 'above'" [class.text-auraRed]="a.condition === 'below'">{{ a.condition }}</span>
                    <span class="font-mono ml-1 text-white">{{ a.targetPrice | number:'1.0-0' }}</span>
                  </div>
                </div>
              </div>
            </div>

          </div>
        </aside>
      </main>

      <!-- Trading Profile Modal -->
      <app-trading-profile-modal 
        *ngIf="showProfileModal" 
        (close)="showProfileModal = false">
      </app-trading-profile-modal>
    </div>
  `,
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {
  bullishStocks: ScreenerResult[] = [];
  alerts: Alert[] = [];
  loading = false;
  loadingAlerts = false;
  showProfileModal = false;

  constructor(private auth: AuthService, private trade: TradeService) {}

  get email() { return this.auth.email; }

  ngOnInit() {
    this.loading = true;
    this.trade.screen({ signal: 'bullish' }).subscribe({
      next: (res) => {
        this.bullishStocks = res.slice(0, 5); // top 5 bullish
        this.loading = false;
      },
      error: () => this.loading = false
    });

    this.loadingAlerts = true;
    this.trade.getAlerts().subscribe({
      next: (res) => {
        this.alerts = res.slice(0, 5); // top 5 alerts
        this.loadingAlerts = false;
      },
      error: () => this.loadingAlerts = false
    });
  }

  logout() { this.auth.logout(); }

  analyze(ticker: string) {
    this.trade.triggerChatWithTicker(ticker);
  }
}

