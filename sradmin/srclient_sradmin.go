package sradmin

import (
	"encoding/base64"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/riferrei/srclient"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"net/http"
)

type SrAdmin struct {
	client *srclient.SchemaRegistryClient
}

type SubjectListingStartedMsg struct {
	subjects chan []Subject
	err      chan error
}

type SubjetListingErrorMsg struct {
	Err error
}

func (msg *SubjectListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case subjects := <-msg.subjects:
		return SubjectsListedMsg{subjects}
	case err := <-msg.err:
		log.Error("Failed to fetch subjects", "err", err)
		return SubjetListingErrorMsg{err}
	}
}

type SubjectsListedMsg struct {
	Subjects []Subject
}

func (s *SrAdmin) ListSubjects() tea.Msg {
	subjectsChan := make(chan []Subject)
	errChan := make(chan error)

	go s.doListSubject(subjectsChan, errChan)

	return SubjectListingStartedMsg{subjectsChan, errChan}
}

type Subject struct {
	Name     string
	Versions []int
}

func (s *SrAdmin) doListSubject(subjectsChan chan []Subject, errChan chan error) {
	maybeIntroduceLatency()

	subjects, err := s.client.GetSubjects()
	if err != nil {
		errChan <- err
		return
	}

	//var wg sync.WaitGroup
	//var mu sync.Mutex
	results := make([][]int, len(subjects))

	//for i, subject := range subjects {
	//
	//	wg.Add(1)
	//
	//	go func(index int, subject string) {
	//		defer wg.Done()
	//		versions, err := s.client.GetSchemaVersions(subject)
	//		if err != nil {
	//			errChan <- fmt.Errorf("failed to get versions for subject %s: %w", subject, err)
	//			return
	//		}
	//		mu.Lock()
	//		results[i] = versions
	//		mu.Unlock()
	//	}(i, subject)
	//
	//}
	//
	//wg.Wait()

	var subjectPtrs []Subject
	for i, str := range subjects {
		ptr := str
		subjectPtrs = append(subjectPtrs, Subject{
			Name:     ptr,
			Versions: results[i],
		})
	}
	subjectsChan <- subjectPtrs
}

type SchemaCreationStartedMsg struct {
	created chan bool
	err     chan error
}

type SchemaCreatedMsg struct{}

func (msg *SchemaCreationStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-msg.created:
		return SchemaCreatedMsg{}
	case err := <-msg.err:
		log.Error("Failed to fetch subjects", "err", err)
		return kadmin.OffsetListingErrorMsg{err}
	}
}

func (s *SrAdmin) CreateSchema(details SubjectCreationDetails) tea.Msg {
	createdChan := make(chan bool)
	errChan := make(chan error)

	go s.doCreateSchema(details, createdChan, errChan)

	return SchemaCreationStartedMsg{
		createdChan,
		errChan,
	}
}

func (s *SrAdmin) doCreateSchema(details SubjectCreationDetails, createdChan chan bool, errChan chan error) {
	maybeIntroduceLatency()
	_, err := s.client.CreateSchema(details.Subject, details.Schema, srclient.Avro)
	if err != nil {
		errChan <- err
	}
	createdChan <- true
}

func NewSrAdmin(ktx *kontext.ProgramKtx) *SrAdmin {
	registry := ktx.Config.ActiveCluster().SchemaRegistry
	client := createHttpClient(registry)
	return &SrAdmin{
		client: srclient.NewSchemaRegistryClient(registry.Url, srclient.WithClient(client)),
	}
}

func createHttpClient(registry *config.SchemaRegistryConfig) *http.Client {
	auth := registry.Username + ":" + registry.Password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Transport: roundTripperWithAuth{
			baseTransport: transport,
			authHeader:    authHeader,
		},
	}
	return client
}

type roundTripperWithAuth struct {
	baseTransport http.RoundTripper
	authHeader    string
}

// RoundTrip adds the Authorization header to every request
func (r roundTripperWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", r.authHeader)
	return r.baseTransport.RoundTrip(req)
}
