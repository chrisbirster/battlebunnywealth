package identity

import "time"

func (s *Service) SetDeviceAttestation(accountID, deviceID, provider, status string, hardwareBacked, productionEligible bool, attestedAt time.Time) (AccountView, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.accountIndexLocked(accountID)
	if idx < 0 { return AccountView{}, ErrUnauthorized }
	for i := range s.state.Accounts[idx].Devices {
		device := &s.state.Accounts[idx].Devices[i]
		if device.ID != deviceID { continue }
		if device.Status == DeviceStatusRevoked { return AccountView{}, ErrUnknownDevice }
		device.AttestationProvider = provider
		device.AttestationStatus = status
		device.HardwareBacked = hardwareBacked
		device.ProductionEligible = productionEligible
		at := attestedAt.UTC()
		device.AttestedAt = &at
		if err := s.store.Save(s.state); err != nil { return AccountView{}, err }
		return view(s.state.Accounts[idx]), nil
	}
	return AccountView{}, ErrUnknownDevice
}
