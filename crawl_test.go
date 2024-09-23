package main

// type CrawlTestSuite struct {
// 	suite.Suite
// 	god    *God
// 	ctx    context.Context
// 	cancel context.CancelFunc
// }
//
// type MockHttpClient struct {
// 	Response *http.Response
// 	Err      error
// }
//
// func (m *MockHttpClient) Do(r *http.Request) (*http.Response, error) {
// 	return m.Response, m.Err
// }
//
// func validUsernameGen(id int) string { return fmt.Sprintf("validUsername%d", id) }
//
// var validUsername = validUsernameGen(1)
//
// func (suite *CrawlTestSuite) SetupTest() {
// 	// suite.ctx, suite.cancel = context.WithCancel(context.Background())
// 	//
// 	// config := &Config{
// 	// 	redisSortedSetKey: "reddit_user_scan_priority_test",
// 	// 	redisAddress:      "localhost:6379",
// 	// 	redisPassword:     "",
// 	// 	redisDB:           0,
// 	// 	dbUser:            "postgres",
// 	// 	dbName:            "persona-prototype-6-test",
// 	// }
// 	//
// 	// client := &MockHttpClient{}
// 	// suite.god = NewGod(config, client)
// 	//
// 	// if err := models.DeleteAllPersona(suite.god.personaDB.DB()); err != nil {
// 	// 	panic(err)
// 	// }
// 	//
// 	// if err := suite.god.sortedSet.Del(suite.god.sortedSet.Key).Err(); err != nil {
// 	// 	panic(err)
// 	// }
// 	//
// 	// // psql
// 	// personaIDStr, err := models.InsertPersona(suite.god.personaDB.DB(), validUsername, models.SOCIAL_NET_REDDIT)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// personaID, _ := uuid.Parse(personaIDStr)
// 	//
// 	// // map
// 	// persona, err := models.FindPersonaByID(suite.god.personaDB.DB(), personaID)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// suite.god.redditUsersMap.Items[validUsername] = &models.PersonaWithActivities{
// 	// 	Persona: persona,
// 	// }
// 	//
// 	// // redis sorted set
// 	// // TODO calculate redis sorted set score based on: priority score, now-last scan, weekly_activity[i]
// 	// err = suite.god.sortedSet.ZAdd(suite.god.sortedSet.Key, redis.Z{
// 	// 	Score:  float64(persona.PriorityScore),
// 	// 	Member: validUsername,
// 	// }).Err()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// }
//
// func (suite *CrawlTestSuite) TearDownTest() {
// 	if err := models.DeleteAllPersona(suite.god.personaDB.DB()); err != nil {
// 		panic(err)
// 	}
//
// 	if err := suite.god.pqueue.Del(suite.god.pqueue.Key).Err(); err != nil {
// 		panic(err)
// 	}
//
// 	defer suite.god.personaDB.Close()
// 	defer suite.god.pqueue.Close()
// }
//
// // TEST: comments: empty username
// // TODO: errors
// // TODO: db error names
// // TODO: request with context
//
// func (suite *CrawlTestSuite) TestCrawl() {
// 	// c := make(chan struct{})
// 	// expectedRedditPersonaFromMap := suite.god.redditUsersMap.Items[validUsername]
// 	// expectedRedditPersonaFromMap.Persona.Thirdparty.About = SuspendedRedditUserAbout()
// 	//
// 	// suspendedRedditUser, _ := json.Marshal(SuspendedRedditUserAbout())
// 	// suite.god.Client = &MockHttpClient{
// 	// 	Response: &http.Response{
// 	// 		StatusCode: http.StatusNotFound,
// 	// 		Body:       io.NopCloser(bytes.NewBuffer(suspendedRedditUser)),
// 	// 	},
// 	// 	Err: nil,
// 	// }
// 	//
// 	// // crawlerID := 1
// 	// go requestTicker(suite.ctx, time.Millisecond)
// 	// go func() {
// 	// 	// crawl(suite.ctx, crawlerID, suite.god)
// 	// 	c <- yes
// 	// }()
// 	//
// 	// time.Sleep(50 * time.Millisecond)
// 	// suite.cancel()
// 	// <-c
// 	// personaFromMap := suite.god.redditUsersMap.Items[validUsername]
// 	//
// 	// assert.NotNil(suite.T(), personaFromMap.Persona.Thirdparty.About)
// 	//
// 	// sortedSetLength := len(suite.god.sortedSet.ZRange(suite.god.sortedSet.Key, 0, -1).Val())
// 	// assert.Equal(suite.T(), 0, sortedSetLength)
// 	// expectedPriorityScore := calculatePriorityScore(expectedRedditPersonaFromMap)
// 	// assert.Equal(suite.T(), expectedPriorityScore, personaFromMap.Persona.PriorityScore)
// 	//
// 	// personaFromDB, _ := models.FindPersonaByID(suite.god.personaDB.DB(), personaFromMap.Persona.ID)
// 	// assert.Equal(suite.T(), expectedRedditPersonaFromMap.Persona.LastScannedAt, personaFromDB.LastScannedAt)
// 	//
// 	// // client Do:
// 	// // 1. with len(contents) > 100
// 	// // 2. with content = 0
// 	// // 3. with spam comments
// 	// // 4. with spam posts
// 	// // 5. with multiple users
// 	// // 6. weekly_activity
// 	// // 7. yearly_activity
// 	// // 8. request context
// 	// // 9. multiple crawlers
// }
//
// func print(a any) {
// 	b, _ := json.MarshalIndent(a, "", " ")
// 	fmt.Println(string(b))
// }
//
// func TestCrawlSuite(t *testing.T) {
// 	suite.Run(t, new(CrawlTestSuite))
// }
