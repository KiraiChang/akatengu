import { getAggregateVersion, appendEvent } from './aggregate';

export interface RateUpdatedPayload {
  currency: string;
  date:     string;
  rate_twd: string;
}

export const createExchangeRate = async (payload: RateUpdatedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<RateUpdatedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'investment.fx_rate_updated',
    payload,
    metadata:         {},
  });
};
