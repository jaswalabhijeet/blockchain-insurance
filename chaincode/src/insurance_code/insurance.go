package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
	"encoding/json"
)

// InsuranceChaincode example simple Chaincode implementation
type InsuranceChaincode struct {
}

//==============================================================================================================================
//	Policy - Defines the structure for a policy object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Policy struct {
	Id		string					`json:"id"`
	Type		string				`json:"type"`
	Details		PolicyDetails		`json:"details"`
	Relations 	PolicyRelations		`json:"relations"`
}

//==============================================================================================================================
//	PolicyDetails - Defines the structure for the PolicyDetails object.
//==============================================================================================================================


type PolicyDetails struct {
	StartDate	string			`json:"startDate"`
	EndDate		string			`json:"endDate"`
	Excess		int				`json:"excess"`
}

//==============================================================================================================================
//	PolicyRelations - Defines the structure for a PolicyRelations object.
//==============================================================================================================================
type PolicyRelations struct {
	Owner		string			`json:"owner"`
	Vehicle		string			`json:"vehicle"`
	Claims		[]string		`json:"claims"`

}

//==============================================================================================================================
//	Claim - Defines the structure for a Claim object.
//==============================================================================================================================
type Claim struct {
	Id		string				`json:"id"`
	Type		string			`json:"type"`
	Details		ClaimDetails	`json:"details"`
	Relations 	ClaimRelations	`json:"relations"`
}

//==============================================================================================================================
//	ClaimDetails - Defines the structure for a ClaimDetails object.
//==============================================================================================================================
type ClaimDetails struct {
	Status		string							`json:"status"`
	Description	string							`json:"description"`
	Incident	ClaimDetailsIncident			`json:"incident"`
	Repair		ClaimDetailsClaimGarageReport	`json:"repair"`
	Settlement	ClaimDetailsSettlement			`json:"settlement"`
}

//==============================================================================================================================
//	ClaimDetailsIncident - Defines the structure for a ClaimDetailsIncident object.
//==============================================================================================================================
type ClaimDetailsIncident struct {
	Date	string	`json:"date"`
	Type	string	`json:"type"`
}

//==============================================================================================================================
//	ClaimDetailsClaimGarageReport - Defines the structure for a ClaimDetailsClaimGarageReport object.
//==============================================================================================================================
type ClaimDetailsClaimGarageReport struct {
	Garage		string	`json:"garage"`
	Estimate	int		`json:"estimate"`
	Actual		int		`json:"actual"`
	WriteOff	bool	`json:"writeOff"`
	Notes		string	`json:"notes"`
}

//==============================================================================================================================
//	ClaimDetailsSettlement - Defines the structure for a ClaimDetailsSettlement object.
//==============================================================================================================================
type ClaimDetailsSettlement struct {
	Decision	string							`json:"decision"`
	Dispute		bool							`json:"dispute"`
	TotalLoss	ClaimDetailsSettlementTotalLoss	`json:"totalLoss"`
	Payments	[]ClaimDetailsSettlementPayment	`json:"payments"`
}

//==============================================================================================================================
//	ClaimDetailsSettlementTotalLoss - Defines the structure for a ClaimDetailsSettlementTotalLoss object.
//==============================================================================================================================
type ClaimDetailsSettlementTotalLoss struct {
	CarValueEstimate	int	`json:"carValueEstimate"`
	CustomerAgreedValue	int	`json:"customerAgreedValue"`
}

//==============================================================================================================================
//	ClaimDetailsSettlementPayment - Defines the structure for a ClaimDetailsSettlementPayment object.
//==============================================================================================================================
type ClaimDetailsSettlementPayment struct {
	RecipientType	string	`json:"recipientType"`
	Recipient	string		`json:"recipient"`
	Amount		int			`json:"amount"`
	Status		string		`json:"status"`
}

//==============================================================================================================================
//	ClaimRelations - Defines the structure for a ClaimRelations object.
//==============================================================================================================================
type ClaimRelations struct {
	RelatedPolicy	string	`json:"relatedPolicy"`
}

//==============================================================================================================================
//	Vehicle - Defines the structure for a Vehicle object.
//==============================================================================================================================
type Vehicle struct {
	Id			string				`json:"id"`
	Type		string				`json:"type"`
	Details		VehicleDetails		`json:"details"`
}

//==============================================================================================================================
//	VehicleDetails - Defines the structure for a VehicleDetails object.
//==============================================================================================================================
type VehicleDetails struct {
	Make			string		`json:"make"`
	Model			string		`json:"model"`
	Registration	string		`json:"registration"`
	Year			string		`json:"year"`
	Mileage			string		`json:"mileage"`
}

//==============================================================================================================================
//	User - Defines the structure for a User object.
//==============================================================================================================================
type User struct {
	Id			string			`json:"id"`
	Type		string			`json:"type"`
	Details		UserDetails		`json:"details"`
	Relations	UserRelations	`json:"relations"`
}

//==============================================================================================================================
//	UserDetails - Defines the structure for a UserDetails object.
//==============================================================================================================================
type UserDetails struct {
	Forename	string		`json:"forename"`
	Surname		string		`json:"surname"`
	Email		string		`json:"email"`
}

//==============================================================================================================================
//	UserRelations - Defines the structure for a UserRelations object.
//==============================================================================================================================
type UserRelations struct {
	RelatedPolicy	string	`json:"relatedPolicy"`
}

//==============================================================================================================================
//	PoliceReport - Defines the structure for a PoliceReport object.
//==============================================================================================================================
type PoliceReport struct {
	Description         string          `json:"description"`
	Location            Coordinates     `json:"coordinates"`
	DriverAtFault       bool            `json:"driver_at_fault"`
}

//==============================================================================================================================
//	Coordinates - Defines the structure for a Coordinates object.
//==============================================================================================================================
type Coordinates struct {
	x float32 `json:"x"`
	y float32 `json:"y"`
}

//==============================================================================================================================
//	 Keys for obtaining the current id for the different domain object types.
//   Ids are incremental so knowing the latest id is useful when querying for
//   all domain objects of a certain type.
//==============================================================================================================================
const   CURRENT_POLICY_ID_KEY      =  "currentPolicyId"
const   CURRENT_CLAIM_ID_KEY   =  "currentClaimId"

//==============================================================================================================================
//	 Prefixes for the different domain object type ids
//==============================================================================================================================
const   POLICY_ID_PREFIX      =  "P"
const   CLAIM_ID_PREFIX   =  "C"

//==============================================================================================================================
//	 Claim Status types - TODO Flesh these out. TODO Following IBM sample, but should/could these be enums?
//==============================================================================================================================
const   STATE_AWAITING_POLICE_REPORT                = "awaiting_police_report"
const   STATE_AWAITING_GARAGE_REPORT                = "awaiting_garage_report"
const   STATE_AWAITING_GARAGE_WORK_CONFIRMATION     = "awaiting_garage_work"
const   STATE_SETTLED  			                    = "settled"
const	STATUS_OPEN									= "open"
const	STATUS_CLOSED								= "closed"

//==============================================================================================================================
//	 Claim Type types - TODO Flesh these out. TODO Following IBM sample, but should these be enums?
//==============================================================================================================================
const   SINGLE_PARTY                =  "single_party"
const   MULTIPLE_PARTIES  			=  "multiple_parties"

//==============================================================================================================================
//	 Settlement Decision types - TODO Flesh these out. TODO Following IBM sample, but should these be enums?
//==============================================================================================================================
const   TOTAL_LOSS                  =  "total_loss"

func main() {
	err := shim.Start(new(InsuranceChaincode))
	if err != nil {
		fmt.Printf("Error starting Insurance chaincode: %s", err)
	}
}

// TODO Set reference data?
func (t *InsuranceChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	stub.PutState(CURRENT_POLICY_ID_KEY, []byte("0"))
	stub.PutState(CURRENT_CLAIM_ID_KEY, []byte("0"))
	return nil, nil
}

// Invoke is the entry point to invoke a chaincode function
func (t *InsuranceChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	caller, caller_affiliation, _ := t.get_caller_data(stub)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "addPolicy" {
		return t.addPolicy(stub, caller, caller_affiliation, args)
	} else if function == "createClaim" {
		return t.createClaim(stub, caller, caller_affiliation, args)
	} else if function == "addPoliceReport" {
		//TODO
	} else if function == "addGarageReport" {
		//TODO
	} else if function == "confirmWork" {
		//TODO
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *InsuranceChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	caller, caller_affiliation, _ := t.get_caller_data(stub)

	// Handle different functions
	if function == "retrieveAllPolicies" {
		return t.retrieveAllPoliciesJSON(stub, caller, caller_affiliation)
	} else if function == "retrieveAllClaims" {
		return t.retrieveAllClaimsJSON(stub, caller, caller_affiliation)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

//=================================================================================================================================
//	 Add Policy  - Creates a Policy object and then saves it to the ledger.
//          args - startDate, endDate, excess, vehicle
//=================================================================================================================================
func (t *InsuranceChaincode) addPolicy(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, args []string) ([]byte, error) {

	fmt.Println("running addPolicy()")

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4 (ActivationDate,ExpiryDate,Excess,VehicleReg)")
	}

	excess, _ := strconv.Atoi(args[2])
	policy := t.newPolicy(t.getNextPolicyId(stub), caller, args[0], args[1], excess, args[3])

	bytes, err := json.Marshal(policy)

	if err != nil { fmt.Printf("addPolicy Error converting policy record: %s", err); return nil, errors.New("Error converting policy record") }

	err = stub.PutState(policy.Id, bytes)

	if err != nil { fmt.Printf("addPolicy: Error storing policy record: %s", err); return nil, errors.New("Error storing policy record") }

	return nil, nil
}

//=================================================================================================================================
//	 newPolicy	-	Constructs a new policy
//=================================================================================================================================
func (t *InsuranceChaincode) newPolicy(id string, owner string, startDate string, endDate string, excess int, vehicleReg string) (Policy) {
	var policy Policy

	policy.Id = id
	policy.Type = "policy"

	policy.Details.StartDate = startDate
	policy.Details.EndDate = endDate
	policy.Details.Excess = excess

	policy.Relations.Owner = owner
	policy.Relations.Vehicle = vehicleReg

	return policy
}

//=================================================================================================================================
//	 createClaim - Creates a Claim object and then saves it to the ledger.
//          args - RelatedPolicy,Description,Date,IncidentType
//=================================================================================================================================
func (t *InsuranceChaincode) createClaim(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, args []string) ([]byte, error) {

	fmt.Println("running createClaim()")

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4 (RelatedPolicy,Description, Date, IncidentType)")
	}

	claim := t.newClaim(t.getNextClaimId(stub), args[0], args[1], args[2], args[3])

	isValid, err := t.checkClaimIsValid(stub, caller, claim);

	if !isValid {
		fmt.Printf("createClaim: Claim invalid: %s", err);
		return nil, errors.New("Claim is invalid");
	}

	bytes, err := json.Marshal(claim)

	if err != nil { fmt.Printf("createClaim Error converting claim record: %s", err); return nil, errors.New("Error converting claim record") }

	err = stub.PutState(claim.Id, bytes)

	if err != nil { fmt.Printf("createClaim: Error storing claim record: %s", err); return nil, errors.New("Error storing claim record") }

	return nil, nil
}

//=================================================================================================================================
//	 newClaim	-	Constructs a new claim
//=================================================================================================================================
func (t *InsuranceChaincode) newClaim(id string, relatedPolicy string, description string, incidentDate string, incidentType string) (Claim) {

	var claim Claim

	claim.Id = id
	claim.Type = "claim"

	claim.Relations.RelatedPolicy = relatedPolicy
	claim.Details.Description = description
	claim.Details.Incident.Date = incidentDate
	claim.Details.Incident.Type = incidentType

	claim.Details.Status = STATUS_OPEN

	return claim
}

//=================================================================================================================================
//	 checkClaimIsValid - Performs checks on the claim to ensure it is valid
//      Returns true if valid, or false and an error if invalid
//=================================================================================================================================
func (t *InsuranceChaincode) checkClaimIsValid(stub shim.ChaincodeStubInterface, caller string, claim Claim) (bool, error) {

	//Check policy exists
	policy, err := t.retrievePolicy(stub, claim.Relations.RelatedPolicy);

	if err != nil {
		fmt.Printf("checkClaimIsValid: Error getting policy with id %s", claim.Relations.RelatedPolicy);
		return false, errors.New("Policy doesnt exist");
	}

	//Check policy owner matches current user
	if policy.Relations.Owner != caller {
		fmt.Printf("checkClaimIsValid: Policy owner is incorrect %s", claim.Relations.RelatedPolicy);
		return false, errors.New("Policy owner incorrect");
	}

	return true, nil;
}

//==============================================================================================================================
//	 retrieveAllPoliciesJSON - Iterates through all policy ids, retreiving each and returning a JSON representation
//==============================================================================================================================
func (t *InsuranceChaincode) retrieveAllPoliciesJSON(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string) ([]byte, error) {
	numberOfPolicies := t.getCurrentPolicyIdNumber(stub)

	result := "["

	for i := 1; i <= numberOfPolicies; i++ {

		policyId := POLICY_ID_PREFIX + strconv.Itoa(i)
		policyJSON, err := t.retrievePolicyJSON(stub, policyId)

		//TODO Check caller has rights to see this policy
		if err != nil {
			fmt.Printf("retrievePolicies: Cannot retrieve policy with id " + policyId+": %s", err)
		} else {
			result += string(policyJSON) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//==============================================================================================================================
//	 retrievePolicy - Gets the state of the data at policyId in the ledger then converts it from the stored
//					JSON into the Policy struct for use in the contract. Returns the Policy struct.
//					Returns empty policy if it errors.
//==============================================================================================================================
func (t *InsuranceChaincode) retrievePolicy(stub shim.ChaincodeStubInterface, policyId string) (Policy, error) {

	var policy Policy

	bytes, err := t.retrievePolicyJSON(stub, policyId)

	if err != nil {	fmt.Printf("retrievePolicy: Cannot read policy: %s", err); return policy, errors.New("retrievePolicy: Cannot read policy")}

	err = json.Unmarshal(bytes, &policy);

	if err != nil {	fmt.Printf("retrievePolicy: Corrupt policy record "+string(bytes)+": %s", err); return policy, errors.New("retrievePolicy: Corrupt policy record"+string(bytes))}

	return policy, nil
}

//==============================================================================================================================
//	 retrievePolicyJSON - Gets the state of the data at policyId in the ledger and returns the JSON representation
//==============================================================================================================================
func (t *InsuranceChaincode) retrievePolicyJSON(stub shim.ChaincodeStubInterface, policyId string) ([]byte, error) {

	bytes, err := stub.GetState(policyId);

	if err != nil {	fmt.Printf("retrievePolicy: Failed to invoke: %s", err); return nil, errors.New("retrievePolicy: Error retrieving policy with policyId = " + policyId) }

	return bytes, nil
}

//==============================================================================================================================
//	 retrieveAllClaimsJSON - Iterates through all claim ids, retreiving each and returning a JSON representation
//==============================================================================================================================
func (t *InsuranceChaincode) retrieveAllClaimsJSON(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string) ([]byte, error) {
	numberOfClaims := t.getCurrentClaimIdNumber(stub)

	result := "["

	for i := 1; i <= numberOfClaims; i++ {

		claimId := CLAIM_ID_PREFIX + strconv.Itoa(i)
		claimJSON, err := t.retrieveClaimJSON(stub, claimId)

		//TODO Check caller has rights to see this claim
		if err != nil {
			fmt.Printf("retrieveAllClaims: Cannot retrieve claim with id " + claimId+": %s", err)
		} else {
			result += string(claimJSON) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//==============================================================================================================================
//	 retrieveClaim - Gets the state of the data at claimId in the ledger then converts it from the stored
//					JSON into the Claim struct for use in the contract. Returns the Claim struct.
//					Returns empty claim if it errors.
//==============================================================================================================================
func (t *InsuranceChaincode) retrieveClaim(stub shim.ChaincodeStubInterface, claimId string) (Claim, error) {

	var claim Claim

	bytes, err := t.retrieveClaimJSON(stub, claimId)

	if err != nil {	fmt.Printf("retrieveClaim: Cannot read claim: %s", err); return claim, errors.New("retrieveClaim: Cannot read claim")}

	err = json.Unmarshal(bytes, &claim);

	if err != nil {	fmt.Printf("retrieveClaim: Corrupt claim record "+string(bytes)+": %s", err); return claim, errors.New("retrieveClaim: Corrupt claim record"+string(bytes))}

	return claim, nil
}

//==============================================================================================================================
//	 retrieveClaimJSON - Gets the state of the data at claimId in the ledger and returns the JSON representation
//==============================================================================================================================
func (t *InsuranceChaincode) retrieveClaimJSON(stub shim.ChaincodeStubInterface, claimId string) ([]byte, error) {

	bytes, err := stub.GetState(claimId);

	if err != nil {	fmt.Printf("retrieveClaimJSON: Failed to invoke: %s", err); return nil, errors.New("retrieveClaimJSON: Error retrieving claim with claimId = " + claimId) }

	return bytes, nil
}

//==============================================================================================================================
//	 ID Functions - The current id of both policies and claims are stored in blockchain state.
//   This value is incremented when a new policy or claim is created.
//   A prefix is added to the id's to differentiate between policies and claims
//==============================================================================================================================
func (t *InsuranceChaincode) getCurrentPolicyIdNumber(stub shim.ChaincodeStubInterface) (int) {
	return t.getCurrentIdNumber(stub, CURRENT_POLICY_ID_KEY);
}

func (t *InsuranceChaincode) getNextPolicyId(stub shim.ChaincodeStubInterface) (string) {

	return t.getNextId(stub, CURRENT_POLICY_ID_KEY, POLICY_ID_PREFIX);
}

func (t *InsuranceChaincode) getCurrentClaimIdNumber(stub shim.ChaincodeStubInterface) (int) {
	return t.getCurrentIdNumber(stub, CURRENT_CLAIM_ID_KEY);
}

func (t *InsuranceChaincode) getNextClaimId(stub shim.ChaincodeStubInterface) (string) {

	return t.getNextId(stub, CURRENT_CLAIM_ID_KEY, CLAIM_ID_PREFIX);
}

func (t *InsuranceChaincode) getNextId(stub shim.ChaincodeStubInterface, idKey string, idPrefix string) (string) {

	currentId := t.getCurrentIdNumber(stub, idKey)

	nextIdNum := strconv.Itoa(currentId + 1)

	stub.PutState(idKey, []byte(nextIdNum))

	return idPrefix + nextIdNum
}

func (t *InsuranceChaincode) getCurrentIdNumber(stub shim.ChaincodeStubInterface, idKey string) (int) {
	bytes, err := stub.GetState(idKey);

	if err != nil { fmt.Printf("getCurrentIdNumber Error getting id %s", err); return -1}

	currentId, err := strconv.Atoi(string(bytes))

	return currentId;
}

//==============================================================================================================================
//	 Security Functions
//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *InsuranceChaincode)  get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error){

	user, err := t.get_username(stub)

	// if err != nil { return "", "", err }

	// ecert, err := t.get_ecert(stub, user);

	// if err != nil { return "", "", err }

	affiliation, err := t.check_affiliation(stub);

	if err != nil { return "", "", err }

	return user, affiliation, nil
}

//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *InsuranceChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

	//username, err := stub.ReadCertAttribute("username");
	//if err != nil {
	//	fmt.Printf("Couldn't get attribute 'username'. Error: %s", err)
	//	return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error())
	//}
    //
	//return string(username), nil
	return "dummy-user", nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *InsuranceChaincode) check_affiliation(stub shim.ChaincodeStubInterface) (string, error) {
	//affiliation, err := stub.ReadCertAttribute("role");
	//if err != nil { return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error()) }
	//return string(affiliation), nil
	return "dummy-affiliation", nil
}
