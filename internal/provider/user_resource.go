package provider

import (
    "context"
    "time"
    "fmt"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var (
    _ resource.Resource = &userResource{}
    _ resource.ResourceWithConfigure = &userResource{}
)

func NewUserResource() resource.Resource {
  return &userResource{}
}

type userResource struct {
  client *mongo.Client
}

type userResourceModel struct {
  User types.String `tfsdk:"user"`
  Password types.String `tfsdk:"password"`
  Db types.String `tfsdk:"db"`
  Roles []userRoleModel `tfsdk:"roles"`
}

type userRoleModel struct {
  Db types.String `tfsdk:"db"`
  Role types.String `tfsdk:"role"`
}

type commandResponse struct {
  OK            int       `bson:"ok"`
  OperationTime time.Time `bson:"operationTime"`
}

type dbUser struct {
  Id         string `bson:"_id"`
  User       string `bson:"user"`
  Db         string `bson:"db"`
  Roles      []dbRole `bson:"roles"`
}

type dbRole struct {
	Role string `bson:"role"`
	Db   string `bson:"db"`
}

type readResponse struct {
  commandResponse `bson:",inline"`
  Users []dbUser `bson:"users"`
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*mongo.Client)

    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *mongo.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )

        return
    }

    r.client = client
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
      Attributes: map[string]schema.Attribute{
        "db": schema.StringAttribute{
          Required: true,
        },
        "user": schema.StringAttribute{
          Required: true,
        },
        "password": schema.StringAttribute{
          Required: true,
          Sensitive: true,
        },
        "roles": schema.ListNestedAttribute{
          Required: true,
          NestedObject: schema.NestedAttributeObject {
            Attributes: map[string]schema.Attribute{
              "db": schema.StringAttribute{
                Required: true,
              },
              "role": schema.StringAttribute{
                Required: true,
              },
            },
          },
        },
      },
    }
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan userResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
      return
    }

    var roles []bson.M
    for _, role := range plan.Roles {
      roles = append(roles, bson.M {"role": role.Role.ValueString(), "db": role.Db.ValueString()})
    }

    userCreateCommand := bson.D {{"createUser", plan.User.ValueString()}, {"pwd", plan.Password.ValueString()}, {"roles", roles}}

    mongoResult := r.client.Database(plan.Db.ValueString()).RunCommand(ctx, userCreateCommand)
    if mongoResult.Err() != nil {
      resp.Diagnostics.AddError(
          "Error creating user",
          "Could not create user, unexpected error: " + mongoResult.Err().Error(),
      )
      return
    }

    var response commandResponse
    err := mongoResult.Decode(&response)
    if err != nil {
      resp.Diagnostics.AddError(
        "Error creating user",
        "Could not create user, unexpected error: " + err.Error(),
      )

      return
    }

    if response.OK != 1 {
      resp.Diagnostics.AddError(
        "Error creating user",
        fmt.Sprintf("Could not create user, unexpected error returned from MongoDB: %d", response.OK))
      return
    }

    // Set state to fully populated data
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
  var state userResourceModel
  diags := req.State.Get(ctx, &state)
  resp.Diagnostics.Append(diags...)
  if resp.Diagnostics.HasError() {
    return
  }

  var usersInfo readResponse
  cmd := bson.D{{Key: "usersInfo", Value: bson.M{
      "user": state.User.ValueString(),
      "db":   state.Db.ValueString(),
  }}}

  err := r.client.Database(state.Db.ValueString()).RunCommand(ctx, cmd).Decode(&usersInfo)
  if err != nil {
    resp.Diagnostics.AddError(
        "Error reading user from MongoDb",
        "Could not retrieve user <" + state.User.ValueString() + "> " + err.Error())
  }


  users := usersInfo.Users
  if len(users) == 0 {
    resp.Diagnostics.AddError(
        "Error reading user from MongoDb",
        "Could not retrieve user <" + state.User.ValueString() + "> " + err.Error())
  }

  user := users[0]
  state.User = types.StringValue(user.User)
  state.Db = types.StringValue(user.Db)

  state.Roles = []userRoleModel{}
  for _, item := range user.Roles {
    state.Roles = append(state.Roles, userRoleModel {
      Db: types.StringValue(item.Db),
      Role: types.StringValue(item.Role),
    })
  }

  diags = resp.State.Set(ctx, &state)
  resp.Diagnostics.Append(diags...)
  if (resp.Diagnostics.HasError()) {
    return
  }
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
