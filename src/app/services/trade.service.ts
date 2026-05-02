import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';

export interface StrategyCardData {
  ticker: string;
  confidence: number;
  entry: number;
  takeProfit: number;
  stopLoss: number;
  rationale: string;
}

export interface TopPick {
  ticker: string;
  currentPrice: number;
  percentChange: number;
}

@Injectable({
  providedIn: 'root'
})
export class TradeService {
  private _chatInputFocus = new Subject<string>();
  chatInputFocus$ = this._chatInputFocus.asObservable();

  constructor(private http: HttpClient) {}

  getTopPicks(): Observable<TopPick[]> {
    return this.http.get<TopPick[]>('/api/top-picks');
  }

  analyzeTicker(ticker: string): Observable<StrategyCardData> {
    return this.http.post<StrategyCardData>('/api/analyze', { ticker });
  }

  triggerChatWithTicker(ticker: string) {
    this._chatInputFocus.next(ticker);
  }
}
