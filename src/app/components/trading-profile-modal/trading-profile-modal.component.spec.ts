import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TradingProfileModalComponent } from './trading-profile-modal.component';

describe('TradingProfileModalComponent', () => {
  let component: TradingProfileModalComponent;
  let fixture: ComponentFixture<TradingProfileModalComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TradingProfileModalComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TradingProfileModalComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
