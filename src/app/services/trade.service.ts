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
  // Technical Analysis (populated by /api/analyze)
  rsi?: number;
  macd?: number;
  macdSignal?: number;
  macdHistogram?: number;
  ma20?: number;
  ma50?: number;
  atr?: number;
  stopLossJustification?: string;
  takeProfitJustification?: string;
}

export interface TopPick {
  ticker: string;
  currentPrice: number;
  percentChange: number;
  volume?: number;
  rsi?: number;
  signal?: string;
}

export interface ConversationTurn {
  role: 'user' | 'aura';
  text: string;
}

export interface ChatResponse {
  text: string;
  strategy?: StrategyCardData;
}

export interface ScreenerResult {
  ticker: string;
  currentPrice: number;
  percentChange: number;
  volume: number;
  rsi: number;
  ma20: number;
  ma50: number;
  macd: number;
  macdSignal: number;
  signal: 'bullish' | 'bearish' | 'neutral';
}

export interface ScreenerFilters {
  minPrice?: number;
  maxPrice?: number;
  minVolume?: number;
  minRSI?: number;
  maxRSI?: number;
  signal?: string;
}

export interface Alert {
  id: number;
  ticker: string;
  condition: 'above' | 'below';
  targetPrice: number;
  isActive: boolean;
  createdAt: string;
}

export interface CreateAlertRequest {
  ticker: string;
  condition: 'above' | 'below';
  targetPrice: number;
}

@Injectable({ providedIn: 'root' })
export class TradeService {
  private _chatInputFocus = new Subject<string>();
  chatInputFocus$ = this._chatInputFocus.asObservable();

  constructor(private http: HttpClient) {}

  getTopPicks(): Observable<TopPick[]> {
    return this.http.get<TopPick[]>('/api/top-picks');
  }

  chat(history: ConversationTurn[], message: string): Observable<ChatResponse> {
    return this.http.post<ChatResponse>('/api/chat', { history, message });
  }

  analyze(ticker: string): Observable<StrategyCardData> {
    return this.http.post<StrategyCardData>('/api/analyze', { ticker });
  }

  screen(filters: ScreenerFilters): Observable<ScreenerResult[]> {
    const params: Record<string, string> = {};
    if (filters.minPrice != null) params['minPrice'] = String(filters.minPrice);
    if (filters.maxPrice != null) params['maxPrice'] = String(filters.maxPrice);
    if (filters.minVolume != null) params['minVolume'] = String(filters.minVolume);
    if (filters.minRSI != null) params['minRSI'] = String(filters.minRSI);
    if (filters.maxRSI != null) params['maxRSI'] = String(filters.maxRSI);
    if (filters.signal) params['signal'] = filters.signal;
    return this.http.get<ScreenerResult[]>('/api/screener', { params });
  }

  getAlerts(): Observable<Alert[]> {
    return this.http.get<Alert[]>('/api/alerts');
  }

  createAlert(req: CreateAlertRequest): Observable<{ id: number; message: string }> {
    return this.http.post<{ id: number; message: string }>('/api/alerts', req);
  }

  deleteAlert(id: number): Observable<void> {
    return this.http.delete<void>(`/api/alerts?id=${id}`);
  }

  triggerChatWithTicker(ticker: string) {
    this._chatInputFocus.next(ticker);
  }
}
